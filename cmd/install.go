package cmd

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/melbahja/goph"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed namestr.service
var serviceFile string

//go:embed namestrt
var runScriptFile string

var home = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return home
}()

var Pass, SSHKey string

const goBinaryPackage = "https://go.dev/dl/go1.20.10.linux-amd64.tar.gz"

const (
	appName       = "namestr"
	passVar       = "pass"
	sshKeyVarName = "sshkey"
	minArgs       = 3
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [<username>@]<domain> <npub> <redirect> [<relays>...]",
	Short: "automatically install namestr on a VPS",
	Long: os.Args[0] + ` install

provision a vps to run namestr with a single command, automatically installs itself onto a remote VPS using the ip address and root password using SSH, configures a systemd unit and starts up the service.
	
for security reasons, the password must be given as an environment variable:

` + getEnvName(passVar) + `=password install [<username>@]<domain> <npub> <redirect> [<relays>...]

if you want to use an SSH key, it can also given as an environment variable ` + getEnvName(sshKeyVarName) + ` in addition to the flag. 

if an SSH key name is given, it will be assumed to be in ` + home + `/.ssh/$` + getEnvName(sshKeyVarName) + `

if the SSH key is encrypted, the environment variable ` + getEnvName(passVar) + ` will be used to unlock it.
`,
	Run: func(cmd *cobra.Command, args []string) {

		s.Log("pass: %s \n", Pass)

		var err error

		if len(args) < minArgs {
			s.Err("minimum %d args required\n", minArgs)
			err = cmd.Help()
			if err != nil {
				s.Fatal("error showing help: %s\n", err)
			}
			os.Exit(1)
		}

		username := "root"
		dom, npub, redir, relays := args[0], args[1], args[2], args[3:]

		if strings.Contains(dom, "@") {
			split := strings.Split(dom, "@")
			if len(split) == 2 && len(split[0]) > 0 && len(split) > 0 {
				s.Log("%v\n", split)
				username, dom = split[0], split[1]
			}
		}
		s.Log("args: %v\n", args)
		s.Log("connecting to %s to set up NIP-05 for:\n", dom)
		s.Log("domain %s\nnpub %s\n redirection %s\n relays %v\n",
			dom, npub, redir, relays)

		var client *goph.Client
		var auth goph.Auth
		if SSHKey != "" {
			// if no pass was given but is required this will fail or
			// the connection will fail
			auth, err = goph.Key(filepath.Join(home, SSHKey), Pass)
			if err != nil {
				s.Fatal("error setting up SSH auth: %s\n", err)
			}
			// if no SSH key was given but a pass was given we use
			// pssword auth
		} else if Pass != "" {
			auth = goph.Password(Pass)
		}
		// one or the other methods must be available or we cannot proceed
		if auth == nil {
			s.Fatal("no authentication method provided to connect to VPS\n")
		}
		// connect to the VPS
		client, err = goph.New(username, dom, auth)
		if err != nil {
			s.Fatal("error setting up ssh connection: %s\n", err)
		}
		defer client.Close()
		s.Log("connected to VPS\n")
	},
}

func init() {

	viper.SetEnvPrefix(appName)
	viper.AutomaticEnv()

	installCmd.PersistentFlags().StringVarP(&SSHKey, "sshkey", "k", "",
		"name of ssh key file to use - eg id_ed25519_sk")

	if p := viper.GetString(passVar); p != "" {
		Pass = p
	}

	if SSHKey == "" {
		if k := viper.GetString(sshKeyVarName); k != "" {
			SSHKey = k
		}
	}

	rootCmd.AddCommand(installCmd)
}

func getEnvName(varName string) string {
	return strings.ToUpper(appName + "_" + varName)
}
