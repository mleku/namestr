package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/autotls"
	"github.com/mleku/appdata"
	"github.com/mleku/ec/schnorr"
	"github.com/mleku/signr/pkg/nostr"
	"github.com/spf13/cobra"
)

type config struct {
	c                         *cobra.Command
	DataDir, message          string
	domain, npub, redirection string
	relays                    []string
	verbose, color            bool
}

var s config
var DataDirPerm os.FileMode = 0700

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "namestr <domain> <npub> <redirection> [<relays>...]",
	Short: "zero configuration NIP-05 DNS<->npub identity verification",
	Long: `namestr
	
a simple tool that lets you spawn a full NIP-05 identity service on a server with only a commandline

it is intended to support the simplest use case for a single user at a single domain where the domain is the name, as described in NIP-05 as _@domain.name for example, and will show domain.name in the client as the identifier
	
namestr automatically uses LetsEncrypt to generate a signed certificate for HTTPS, and serves up a https://domain.name/.well_known/nostr.json based on the parameters given in the commandline. the email field is mandatory.
	
optionally a list of relays can be named, only their domain name is required, the necessary prefixes will be appended`,
	Run: func(cmd *cobra.Command, args []string) {

		s.c = cmd

		if len(args) == 0 {
			s.c.Help()
			os.Exit(0)
		}

		if len(args) < 2 {
			s.c.Help()
			s.Fatal("at least domain name and npub is required\n")
		}
		s.domain, s.npub, s.redirection, s.relays = args[0], args[1], args[2], args[3:]
		s.Log("domain %v\n", s.domain)
		s.Log("npub %v\n", s.npub)
		s.Log("relays %v\n", s.relays)

		pk, err := nostr.NpubToPublicKey(s.npub)
		if err != nil {
			s.Fatal("%s\n", err)
		}
		pkb := schnorr.SerializePubKey(pk)
		pkHex := fmt.Sprintf("%x", pkb)
		s.Log("nostr pubkey hex: %s\n", pkHex)
		// generate response test.
		jsonText := `{ 
  "names": { 
    "_": "` + pkHex +
			`" 
  }, 
  "relays": {
    "` + pkHex + `": [
`
		rl := len(s.relays) - 1
		for i, rel := range s.relays {
			if i < rl {
				jsonText +=
					`      ` + `"wss://` + rel + `",
`
			} else {
				jsonText +=
					`      ` + `"wss://` + rel + `"
`
			}
		}

		jsonText +=
			`    ]
  }
}	
`
		s.message = jsonText
		s.Log("text that will be sent in response to requests:\n%s\n",
			s.message)
		s.DataDir = appdata.GetDataDir("namestr", false)
		_, exists, err := CheckFileExists(s.DataDir)
		if err != nil {
			s.Fatal("error checking if datadir exists: %s", err)
		}
		if !exists {
			s.Info("First run: Creating namestr data directory at %s\n\n",
				s.DataDir)
			if err = os.MkdirAll(s.DataDir, DataDirPerm); err != nil {
				s.Fatal("unable to create data dir, cannot proceed: %s",
					err)
			}
		}
		s.Log("starting up server\n")
		autotls.Run(s, s.domain)
		s.Log("finished running server\n")
	},
}

func CheckFileExists(name string) (fi os.FileInfo, exists bool, err error) {
	exists = true
	if fi, err = os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			exists = false
			err = nil
		}
	}
	return
}

func (s config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Log("%s\n", r.RequestURI)
	if r.RequestURI == "/.well-known/nostr.json?name=_" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(s.message))
	} else {
		http.Redirect(w, r, s.redirection, http.StatusSeeOther)
	}
}

// Execute adds all child commands to the root command and sets flags
// appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.BoolVarP(&s.verbose, "verbose", "v", false,
		"prints more things")
	pf.BoolVarP(&s.color, "color", "c", false,
		"prints color things")
}
