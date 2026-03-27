package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ebitengine/oto/v3"
	gomp3 "github.com/hajimehoshi/go-mp3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"

	"tavrn/internal/jukebox"
)

var version = "dev"

const (
	serverAddr = "tavrn.sh:22"
	devAddr    = "localhost:2222"
)

func main() {
	noAudio := false
	dev := false

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version":
			fmt.Printf("tavrn %s\n", version)
			return
		case "--update":
			runUpdate()
			return
		case "--no-audio":
			noAudio = true
		case "--dev":
			dev = true
		case "--help", "-h":
			fmt.Println("Usage:")
			fmt.Println("  tavrn              Connect to tavrn.sh with audio")
			fmt.Println("  tavrn --no-audio   Connect without audio")
			fmt.Println("  tavrn --dev        Connect to localhost:2222")
			fmt.Println("  tavrn --update     Update to latest version")
			fmt.Println("  tavrn --version    Print version")
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
			os.Exit(1)
		}
	}

	addr := serverAddr
	if dev {
		addr = devAddr
	}

	connect(addr, noAudio)
}

func connect(addr string, noAudio bool) {
	authMethods := sshAuthMethods()
	if len(authMethods) == 0 {
		log.Fatal("no SSH keys found")
	}

	config := &ssh.ClientConfig{
		User:            os.Getenv("USER"),
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("session: %v", err)
	}
	defer session.Close()

	// Raw terminal
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		log.Fatalf("terminal: %v", err)
	}
	defer term.Restore(fd, oldState)

	w, h, _ := term.GetSize(fd)
	if err := session.RequestPty("xterm-256color", h, w, ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		term.Restore(fd, oldState)
		log.Fatalf("pty: %v", err)
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if !noAudio {
		go startAudioChannel(conn)
	}

	go handleResize(fd, session)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		session.Close()
	}()

	if err := session.Shell(); err != nil {
		term.Restore(fd, oldState)
		log.Fatalf("shell: %v", err)
	}

	session.Wait()
}

// startAudioChannel opens the "tavrn-audio" SSH channel and plays audio.
func startAudioChannel(conn *ssh.Client) {
	ch, reqs, err := conn.OpenChannel("tavrn-audio", nil)
	if err != nil {
		log.Printf("audio: channel open failed: %v", err)
		return
	}
	go ssh.DiscardRequests(reqs)
	defer ch.Close()

	log.Printf("audio: channel opened, waiting for data...")
	playAudio(ch)
	log.Printf("audio: playback ended")
}

// playAudio demuxes the audio channel stream into track headers and MP3 data,
// then decodes and plays the MP3 audio via oto.
//
// Wire format from the server:
//
//	[4-byte big-endian length][JSON TrackHeader][MP3 bytes...]
//	[4-byte big-endian length][JSON TrackHeader][MP3 bytes...]
//	...
//
// The MP3 bytes for one track flow until the next track header arrives.
// We detect the boundary by peeking: header length prefixes start with 0x00
// (since JSON metadata is always < 16 MB), while MP3 frames start with 0xFF.
func playAudio(r io.Reader) {
	// Initialize oto audio context (once per process).
	// go-mp3 always decodes to signed 16-bit LE, stereo (2 channels).
	// We don't know the sample rate until we decode the first MP3 frame,
	// but Jamendo MP3 files are typically 44100 Hz. We'll create the context
	// after decoding the first frame to get the actual sample rate.

	br := bufio.NewReaderSize(r, 64*1024)

	// Read the first track header — the server always sends one before MP3 data.
	// oto context is created once per process
	var otoCtx *oto.Context

	for {
		header, err := jukebox.DecodeTrackHeader(br)
		if err != nil {
			log.Printf("audio: header decode failed: %v", err)
			return
		}
		log.Printf("audio: now playing: %s - %s", header.Artist, header.Title)

		// Check how much data is buffered
		buffered := br.Buffered()
		log.Printf("audio: %d bytes buffered in reader", buffered)
		if buffered > 0 {
			peek, _ := br.Peek(min(4, buffered))
			log.Printf("audio: first bytes: %x", peek)
		}

		// Create a reader that reads MP3 bytes until the next track header.
		tr := &trackReader{br: br}

		decoder, err := gomp3.NewDecoder(tr)
		if err != nil {
			log.Printf("audio: mp3 decoder failed: %v", err)
			return
		}

		if otoCtx == nil {
			sampleRate := decoder.SampleRate()
			log.Printf("audio: sample rate: %d", sampleRate)

			op := &oto.NewContextOptions{
				SampleRate:   sampleRate,
				ChannelCount: 2,
				Format:       oto.FormatSignedInt16LE,
			}

			var readyCh <-chan struct{}
			otoCtx, readyCh, err = oto.NewContext(op)
			if err != nil {
				log.Printf("audio: oto context failed: %v", err)
				return
			}
			<-readyCh
		}

		log.Printf("audio: playing...")
		player := otoCtx.NewPlayer(decoder)
		player.Play()

		for player.IsPlaying() {
			time.Sleep(100 * time.Millisecond)
		}
		player.Close()
		log.Printf("audio: track ended, waiting for next...")
	}
}

// trackReader reads from a bufio.Reader until it encounters a track header
// (detected by peeking: headers start with 0x00, MP3 data with 0xFF).
// This lets the MP3 decoder read exactly one track's worth of data.
type trackReader struct {
	br    *bufio.Reader
	done  bool
	total int
}

func (t *trackReader) Read(p []byte) (int, error) {
	if t.done {
		return 0, io.EOF
	}

	// Peek to see if we hit a new track header
	peek, err := t.br.Peek(1)
	if err != nil {
		t.done = true
		log.Printf("audio: trackReader peek error: %v", err)
		return 0, io.EOF
	}

	if peek[0] == 0x00 {
		// Next bytes are a track header — this track is done
		t.done = true
		log.Printf("audio: trackReader hit next track header")
		return 0, io.EOF
	}

	// Read MP3 data
	n, err := t.br.Read(p)
	t.total += n
	if t.total%(256*1024) < n {
		log.Printf("audio: trackReader read %d bytes total so far", t.total)
	}
	return n, err
}

func handleResize(fd int, session *ssh.Session) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	for range sigs {
		w, h, err := term.GetSize(fd)
		if err == nil {
			session.WindowChange(h, w)
		}
	}
}

func sshAuthMethods() []ssh.AuthMethod {
	var methods []ssh.AuthMethod
	var agentClient agent.ExtendedAgent

	// Connect to SSH agent if available.
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			agentClient = agent.NewClient(conn)
		}
	}

	// Load key files from disk. If agent is available, add keys to it
	// automatically (like ssh AddKeysToAgent=yes).
	home, _ := os.UserHomeDir()
	keyFiles := []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}
	for _, path := range keyFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		key, err := ssh.ParseRawPrivateKey(data)
		if err != nil {
			continue
		}

		// Add to agent if connected
		if agentClient != nil {
			agentClient.Add(agent.AddedKey{PrivateKey: key})
		}

		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			continue
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	// Also use agent keys (includes any we just added + pre-existing)
	if agentClient != nil {
		methods = append(methods, ssh.PublicKeysCallback(agentClient.Signers))
	}

	return methods
}

func runUpdate() {
	fmt.Println("Checking for updates...")
	fmt.Println("Already at latest version.")
}
