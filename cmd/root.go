package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
	"github.com/mleku/appdata"
	"github.com/mleku/ec/schnorr"
	"github.com/mleku/signr/pkg/nostr"
	"github.com/spf13/cobra"
)

type config struct {
	c              *cobra.Command
	DataDir        string
	verbose, color bool
}

var s config
var DataDirPerm os.FileMode = 0700

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "namestr <domain> <npub> <email> [<relays>...]",
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
		domain, npub, email, relays := args[0], args[1], args[2], args[3:]
		s.Log("domain %v\n", domain)
		s.Log("npub %v\n", npub)
		s.Log("relays %v\n", relays)

		pk, err := nostr.NpubToPublicKey(npub)
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
		rl := len(relays) - 1
		for i, rel := range relays {
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
		handlr := Handler{jsonText}
		s.Log("text that will be sent in response to requests:\n%s\n",
			handlr.message)
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
		var (
			// the structure that handles reloading the certificate
			certReloader *simplecert.CertReloader
			numRenews    int
			ctx, cancel  = context.WithCancel(context.Background())

			tlsConf = tlsconfig.
				NewServerTLSConfig(tlsconfig.TLSModeServerStrict)
			makeServer = func() *http.Server {
				return &http.Server{
					Addr:      ":443",
					Handler:   handlr,
					TLSConfig: tlsConf,
				}
			}
			srv = makeServer()
			cfg = simplecert.Default
		)
		s.Log("lfg\n")
		cfg.Domains = []string{domain}
		cfg.CacheDir = s.DataDir
		cfg.SSLEmail = email
		cfg.HTTPAddress = ""
		cfg.WillRenewCertificate = func() { cancel() }
		cfg.DidRenewCertificate = func() {
			numRenews++
			ctx, cancel = context.WithCancel(context.Background())
			srv = makeServer()
			certReloader.ReloadNow()
			go s.serve(ctx, srv)
		}
		certReloader, err = simplecert.Init(cfg, func() { os.Exit(0) })
		if err != nil {
			s.Fatal("simplecert init failed: %s\n", err)
		}
		go http.ListenAndServe(":80", http.HandlerFunc(simplecert.Redirect))
		log.Println("will serve at: https://" + cfg.Domains[0])
		s.serve(ctx, srv)
		select {}
	},
}

func (s *config) serve(ctx context.Context, srv *http.Server) {
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil &&
			err != http.ErrServerClosed {

			s.Fatal("listen: %+s\n", err)
		}
	}()
	s.Info("server started")
	<-ctx.Done()
	s.Info("server stopped")
	ctxShutDown, cancel := context.
		WithTimeout(context.Background(), 5*time.Second)
	defer func() { cancel() }()
	err := srv.Shutdown(ctxShutDown)
	if err == http.ErrServerClosed {
		s.Log("server exited properly")
	} else if err != nil {
		s.Err("server encountered an error on exit: %+s\n", err)
	}
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

type Handler struct {
	message string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h.message))
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
