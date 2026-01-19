package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/kong"
	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/output"
)

const Version = "0.2.0"

func New() *CLI {
	return &CLI{}
}

type CLI struct {
	Globals Globals `kong:"embed"`

	Auth     AuthCmd     `kong:"cmd,help='Authentication and cookies.'"`
	Search   SearchCmd   `kong:"cmd,help='Search Spotify.'"`
	Track    TrackCmd    `kong:"cmd,help='Track operations.'"`
	Album    AlbumCmd    `kong:"cmd,help='Album operations.'"`
	Artist   ArtistCmd   `kong:"cmd,help='Artist operations.'"`
	Playlist PlaylistCmd `kong:"cmd,help='Playlist operations.'"`
	Show     ShowCmd     `kong:"cmd,help='Show operations.'"`
	Episode  EpisodeCmd  `kong:"cmd,help='Episode operations.'"`

	Play    PlayCmd    `kong:"cmd,help='Start playback.'"`
	Pause   PauseCmd   `kong:"cmd,help='Pause playback.'"`
	Next    NextCmd    `kong:"cmd,help='Skip to next.'"`
	Prev    PrevCmd    `kong:"cmd,help='Skip to previous.'"`
	Seek    SeekCmd    `kong:"cmd,help='Seek within track.'"`
	Volume  VolumeCmd  `kong:"cmd,help='Set volume.'"`
	Shuffle ShuffleCmd `kong:"cmd,help='Toggle shuffle.'"`
	Repeat  RepeatCmd  `kong:"cmd,help='Set repeat mode.'"`
	Status  StatusCmd  `kong:"cmd,help='Playback status.'"`

	Queue   QueueCmd   `kong:"cmd,help='Queue operations.'"`
	Library LibraryCmd `kong:"cmd,help='Library operations.'"`
	Device  DeviceCmd  `kong:"cmd,help='Playback devices.'"`
}

type Globals struct {
	Config             string           `help:"Config file path." env:"SPOGO_CONFIG"`
	Profile            string           `help:"Profile name." env:"SPOGO_PROFILE"`
	Timeout            time.Duration    `help:"HTTP timeout." env:"SPOGO_TIMEOUT" default:"10s"`
	Market             string           `help:"Market country code." env:"SPOGO_MARKET"`
	Language           string           `help:"Language/locale." env:"SPOGO_LANGUAGE"`
	Device             string           `help:"Device name or id." env:"SPOGO_DEVICE"`
	Engine             string           `help:"Engine (auto|web|connect|applescript)." env:"SPOGO_ENGINE"`
	ConnectUserAgent   string           `help:"Connect client User-Agent override." env:"SPOGO_CONNECT_USER_AGENT"`
	ConnectAppPlatform string           `help:"Connect client app-platform override." env:"SPOGO_CONNECT_APP_PLATFORM"`
	ConnectDeviceName  string           `help:"Connect client device name override." env:"SPOGO_CONNECT_DEVICE_NAME"`
	ConnectDeviceModel string           `help:"Connect client device model override." env:"SPOGO_CONNECT_DEVICE_MODEL"`
	VerifyPlayback     time.Duration    `help:"After play/transfer, poll playback for this duration; warn if stuck (progress_ms=0 or missing item/device)." env:"SPOGO_VERIFY_PLAYBACK" default:"0s"`
	VerifyPlaybackFail bool             `help:"Exit non-zero if playback verification fails." env:"SPOGO_VERIFY_PLAYBACK_FAIL"`
	JSON               bool             `help:"JSON output." env:"SPOGO_JSON"`
	Plain              bool             `help:"Plain output." env:"SPOGO_PLAIN"`
	NoColor            bool             `help:"Disable color output." env:"SPOGO_NO_COLOR"`
	Quiet              bool             `short:"q" help:"Quiet output." env:"SPOGO_QUIET"`
	Verbose            bool             `short:"v" help:"Verbose output." env:"SPOGO_VERBOSE"`
	Debug              bool             `short:"d" help:"Debug output." env:"SPOGO_DEBUG"`
	NoInput            bool             `help:"Disable prompts." env:"SPOGO_NO_INPUT"`
	Version            kong.VersionFlag `help:"Print version."`
}

func (g Globals) Settings() (app.Settings, error) {
	format, err := outputFormat(g.JSON, g.Plain)
	if err != nil {
		return app.Settings{}, err
	}
	return app.Settings{
		ConfigPath:         g.Config,
		Profile:            g.Profile,
		Timeout:            g.Timeout,
		Market:             g.Market,
		Language:           g.Language,
		Device:             g.Device,
		Engine:             g.Engine,
		ConnectUserAgent:   g.ConnectUserAgent,
		ConnectAppPlatform: g.ConnectAppPlatform,
		ConnectDeviceName:  g.ConnectDeviceName,
		ConnectDeviceModel: g.ConnectDeviceModel,
		VerifyPlayback:     g.VerifyPlayback,
		VerifyPlaybackFail: g.VerifyPlaybackFail,
		Format:             format,
		NoColor:            g.NoColor,
		Quiet:              g.Quiet,
		Verbose:            g.Verbose,
		Debug:              g.Debug,
		NoInput:            g.NoInput,
	}, nil
}

func outputFormat(jsonFlag, plainFlag bool) (output.Format, error) {
	if jsonFlag && plainFlag {
		return "", errors.New("--json and --plain are mutually exclusive")
	}
	if jsonFlag {
		return output.FormatJSON, nil
	}
	if plainFlag {
		return output.FormatPlain, nil
	}
	return output.FormatHuman, nil
}

func VersionVars() map[string]string {
	return map[string]string{
		"version": Version,
	}
}

func Usage() string {
	return fmt.Sprintf("spogo %s", Version)
}
