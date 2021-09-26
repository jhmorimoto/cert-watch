package util

//goland:noinspection ALL
import (
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"os"
	"path/filepath"
)

func ProcessScp(cw *certwatchv1.CertWatcher, credentialSecret v1.Secret, certFilesDir string) error {
	var err error
	var scpClient scp.Client
	var scpClientConfig ssh.ClientConfig

	username, ok := credentialSecret.Data["username"]
	if !ok {
		return fmt.Errorf("missing credential value from %s/%s: username", credentialSecret.Namespace, credentialSecret.Name)
	}

	if cw.Spec.Actions.Scp.AuthType == "password" {
		password, ok := credentialSecret.Data["password"]
		if !ok {
			return fmt.Errorf("missing credential value from %s/%s: password", credentialSecret.Namespace, credentialSecret.Name)
		}
		scpClientConfig, err = auth.PasswordKey(string(username), string(password), ssh.InsecureIgnoreHostKey())
		if err != nil {
			return fmt.Errorf("error creating ssh client configuration: %s", err.Error())
		}
	} else if cw.Spec.Actions.Scp.AuthType == "key" {
		privatekeyPassphrase := credentialSecret.Data["passphrase"]
		privatekey, ok := credentialSecret.Data["key"]
		if !ok {
			return fmt.Errorf("missing credential value from %s/%s: key", credentialSecret.Namespace, credentialSecret.Name)
		}
		workspacedir, err := os.MkdirTemp("", "certwatch_scp")
		defer os.RemoveAll(workspacedir)
		if err != nil {
			return fmt.Errorf("error creating workspace directory %s: %s", workspacedir, err.Error())
		}
		privatekeyFilename := filepath.Join(workspacedir, "ssh.key")
		err = ioutil.WriteFile(privatekeyFilename, privatekey, 0600)
		if err != nil {
			return fmt.Errorf("error exporting ssh private key to file %s/%s: %s", credentialSecret.Namespace, credentialSecret.Name, err.Error())
		}
		if string(privatekeyPassphrase) != "" {
			scpClientConfig, err = auth.PrivateKeyWithPassphrase(string(username), privatekeyPassphrase, privatekeyFilename, ssh.InsecureIgnoreHostKey())
		} else {
			scpClientConfig, err = auth.PrivateKey(string(username), privatekeyFilename, ssh.InsecureIgnoreHostKey())
		}
	}
	scpClientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	if cw.Spec.Actions.Scp.Port == 0 {
		cw.Spec.Actions.Scp.Port = 22
	}
	remoteHostAddr := fmt.Sprintf("%s:%d", cw.Spec.Actions.Scp.Hostname, cw.Spec.Actions.Scp.Port)

	for _, scpFile := range cw.Spec.Actions.Scp.Files {
		scpClient = scp.NewClient(remoteHostAddr, &scpClientConfig)
		err = scpClient.Connect()
		if err != nil {
			return fmt.Errorf("error connecting to ssh remote host %s - %s", remoteHostAddr, err.Error())
		}

		if scpFile.Mode == "" {
			scpFile.Mode = "0600"
		}
		certFile, err := os.Open(filepath.Join(certFilesDir, scpFile.Name))
		if err != nil {
			return fmt.Errorf("error opening certifiate file %s: %s", scpFile.Name, err.Error())
		}
		err = scpClient.CopyFile(certFile, filepath.Join(scpFile.RemotePath, scpFile.Name), scpFile.Mode)
		if err != nil {
			return fmt.Errorf("error copying certifiate file %s: %s", scpFile.Name, err.Error())
		}
		scpClient.Close()
	}

	return nil
}
