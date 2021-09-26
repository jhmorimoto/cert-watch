package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	apicorev1 "k8s.io/api/core/v1"
)

// Export certificates from the Secret, namely tls.key and tls.crt, into a
// temporary working directory. The path for this temporary directory will be
// returned and is expected to be removed after CertWatcher Action processing is
// completed.
//
// Along with the original tls.key and tls.crt files, additional converted
// versions of the same files will be included. The full list:
//
// - tls.key
// - tls.crt
// - tls.p12         (tls.key and tls.crt included)
// - tls.crt.p12     (tls.crt included)
//
// And also zipped versions of each one:
//
// - tls.key.zip     (tls.key zipped)
// - tls.crt.zip     (tls.crt zipped)
// - tls.zip         (tls.crt and tls.key zipped)
// - tls.p12.zip     (tls.p12 zipped)
// - tls.crt.p12.zip (tls.crt.p12 zipped)
// - tls.all.zip     (tls.key, tls.crt, tls.p12 and tls.crt.p12 zipped)
//
// The "tls" filename prefix is determined by the filenamesPrefix argument.
//
// If a pkcs12Password is provided, *.p12 files will be created with that password.
//
// If a zipFilesPassword is provided, *.zip files will be created with that
// password.
//
// If the function finishes successfully, creating all files, the path of the
// temporary working directory followed by a nil error is returned. Otherwise,
// an empty string followed by the error is returned.
//
func CreateCertificateFiles(secret *apicorev1.Secret, filenamesPrefix string, zipFilesPassword string, pkcs12Password string) (string, error) {
	var err error
	var secretname string = fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)
	var filename string
	var cmd *exec.Cmd

	if filenamesPrefix == "" {
		filenamesPrefix = "tls"
	}

	workspacedir, err := os.MkdirTemp("", "certwatch")
	if err != nil {
		return workspacedir, fmt.Errorf("CreateCertificateFiles cannot create temporary directory: %s", err.Error())
	}
	// defer os.Remove(workspacedir)

	tlsKey, ok := secret.Data["tls.key"]
	if !ok {
		return workspacedir, fmt.Errorf("secret %s does not have value for tls.key", secretname)
	}

	tlsCrt, ok := secret.Data["tls.crt"]
	if !ok {
		return workspacedir, fmt.Errorf("secret %s does not have value for tls.crt", secretname)
	}

	filename = filepath.Join(workspacedir, filenamesPrefix+".key")
	err = os.WriteFile(filename, []byte(tlsKey), 0600)
	if err != nil {
		return workspacedir, fmt.Errorf("CreateCertificateFiles cannot create %s.key: %s", filenamesPrefix, err.Error())
	}

	filename = filepath.Join(workspacedir, filenamesPrefix+".crt")
	err = os.WriteFile(filename, []byte(tlsCrt), 0600)
	if err != nil {
		return workspacedir, fmt.Errorf("CreateCertificateFiles cannot %s.crt: %s", filenamesPrefix, err.Error())
	}

	// openssl pkcs12 -export -out tls.p12 -in tls.crt -inkey tls.key -passin pass:changeit

	cmd = exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-out", workspacedir+"/"+filenamesPrefix+".p12",
		"-in", workspacedir+"/"+filenamesPrefix+".crt",
		"-inkey", workspacedir+"/"+filenamesPrefix+".key",
		"-passout", "pass:"+pkcs12Password,
	)
	// fmt.Println(cmd)
	cmdoutput, err := cmd.CombinedOutput()
	if err != nil {
		return workspacedir, fmt.Errorf("CreateCertificateFiles cannot create "+filenamesPrefix+".p12: %s\n%s", err.Error(), cmdoutput)
	}

	cmd = exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-nokeys",
		"-out", workspacedir+"/"+filenamesPrefix+".crt.p12",
		"-in", workspacedir+"/"+filenamesPrefix+".crt",
		"-passout", "pass:"+pkcs12Password,
	)
	// fmt.Println(cmd)
	cmdoutput, err = cmd.CombinedOutput()
	if err != nil {
		return workspacedir, fmt.Errorf("CreateCertificateFiles cannot create "+filenamesPrefix+".crt.p12: %s\n%s", err.Error(), cmdoutput)
	}

	err = zipFiles(
		workspacedir,
		filenamesPrefix+".key.zip",
		zipFilesPassword,
		[]string{
			filenamesPrefix + ".key",
		},
	)
	if err != nil {
		return workspacedir, err
	}

	err = zipFiles(
		workspacedir,
		filenamesPrefix+".crt.zip",
		zipFilesPassword,
		[]string{
			filenamesPrefix + ".crt",
		},
	)
	if err != nil {
		return workspacedir, err
	}

	err = zipFiles(
		workspacedir,
		filenamesPrefix+".zip",
		zipFilesPassword,
		[]string{
			filenamesPrefix + ".key",
			filenamesPrefix + ".crt",
		},
	)
	if err != nil {
		return workspacedir, err
	}

	err = zipFiles(
		workspacedir,
		filenamesPrefix+".p12.zip",
		zipFilesPassword,
		[]string{
			filenamesPrefix + ".p12",
		},
	)
	if err != nil {
		return workspacedir, err
	}

	err = zipFiles(
		workspacedir,
		filenamesPrefix+".crt.p12.zip",
		zipFilesPassword,
		[]string{
			filenamesPrefix + ".crt.p12",
		},
	)
	if err != nil {
		return workspacedir, err
	}

	err = zipFiles(
		workspacedir,
		filenamesPrefix+".all.zip",
		zipFilesPassword,
		[]string{
			filenamesPrefix + ".key",
			filenamesPrefix + ".crt",
			filenamesPrefix + ".p12",
			filenamesPrefix + ".crt.p12",
		},
	)
	if err != nil {
		return workspacedir, err
	}

	return workspacedir, err
}

func zipFiles(workspacedir string, zipfilename string, password string, files []string) error {
	var cmdargs []string
	var cmd *exec.Cmd
	if password != "" {
		cmdargs = append(cmdargs, "-P", password)
	}
	cmdargs = append(cmdargs, zipfilename)
	cmdargs = append(cmdargs, files...)
	cmd = exec.Command("zip", cmdargs...)
	cmd.Dir = workspacedir
	// fmt.Println(cmd)
	cmdoutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CreateCertificateFiles cannot create %s: %s\n%s", zipfilename, err.Error(), cmdoutput)
	}
	return nil
}
