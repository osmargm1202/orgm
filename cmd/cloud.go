package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Cloud commands: sync config files with R2


func cloudAppConfigPathLocal() (string, error) {
	base, err := resolveConfigDir()
	if err != nil { return "", err }
	return filepath.Join(base, "config.toml"), nil
}




func CloudPullAppConfig() error {
	ctx := context.Background()
	baseURL := viper.GetString("cloudflare.bucket.orgm-privado.url")
	token := viper.GetString("cloudflare.bucket.orgm-privado.token")
	if baseURL == "" { return fmt.Errorf("missing R2 base URL in configuration") }
	if token == "" { return fmt.Errorf("missing R2 token in configuration") }
	url := baseURL
	if url[len(url)-1] == '/' { url = url[:len(url)-1] }
	url = url + "/config.toml"
	data, err := r2HTTPGet(ctx, url, token)
	if err != nil { return err }
	local, err := cloudAppConfigPathLocal()
	if err != nil { return err }
	return SaveBytes(local, data)
}

func CloudPushAppConfig() error {
	ctx := context.Background()
	bucketKey := "orgm_privado" // Use the bucket key from keys.toml
	key := "config.toml"
	local, err := cloudAppConfigPathLocal()
	if err != nil { return err }
	data, err := os.ReadFile(local)
	if err != nil { return err }
	return r2S3Put(ctx, bucketKey, key, data)
}





// config command group
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [editor]",
		Short: "Init/update/edit config.toml synced with R2",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Printf("%s\n", inputs.TitleStyle.Render("config options"))
				fmt.Println(" - init     : download config.toml")
				fmt.Println(" - update   : upload config.toml")
				fmt.Println(" - [editor] : edit config.toml locally (e.g., config nano)")
				return nil
			}
			// Dynamic editor/init/update
			verb := args[0]
			if verb == "init" {
				if err := CloudPullAppConfig(); err != nil { return err }
				fmt.Printf("%s\n", inputs.SuccessStyle.Render("config.toml downloaded"))
				return nil
			}
			if verb == "update" {
				if err := CloudPushAppConfig(); err != nil { return err }
				fmt.Printf("%s\n", inputs.SuccessStyle.Render("config.toml uploaded"))
				return nil
			}
			// Edit local file
			local, err := cloudAppConfigPathLocal()
			if err != nil { return err }
			ed := exec.Command(verb, local)
			ed.Stdin, ed.Stdout, ed.Stderr = os.Stdin, os.Stdout, os.Stderr
			return ed.Run()
		},
	}
	return cmd
}