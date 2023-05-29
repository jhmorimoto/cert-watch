package util

import (
	"errors"
	"strings"

	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	"github.com/magiconair/properties"
	mail "github.com/xhit/go-simple-mail/v2"
)

func ProcessEmail(cw *certwatchv1.CertWatcher, certFilesDir string, emailConfiguration *properties.Properties) error {
	var err error

	if emailConfiguration == nil {
		return errors.New("email not configured")
	}

	server := mail.NewSMTPClient()
	server.Host = emailConfiguration.MustGetString("host")
	server.Port = emailConfiguration.MustGetInt("port")
	server.Username = emailConfiguration.GetString("username", "")
	server.Password = emailConfiguration.GetString("password", "")

	sEncryptionType := emailConfiguration.GetString("encryption", "")
	switch sEncryptionType {
	case "SSL":
		server.Encryption = mail.EncryptionSSL
	case "TLS":
		server.Encryption = mail.EncryptionTLS
	case "SSLTLS":
		server.Encryption = mail.EncryptionSSLTLS
	case "STARTTLS":
		server.Encryption = mail.EncryptionSTARTTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	var from string
	if cw.Spec.Actions.Email.From != "" {
		from = cw.Spec.Actions.Email.From
	} else {
		from = emailConfiguration.MustGetString("from")
	}

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from)
	// var emailString string
	for _, emailString := range strings.Split(cw.Spec.Actions.Email.To, ",") {
		email.AddTo(emailString)
	}
	if cw.Spec.Actions.Email.Cc != "" {
		// email.AddCc(cw.Spec.Actions.Email.Cc)
		for _, emailString := range strings.Split(cw.Spec.Actions.Email.Cc, ",") {
			email.AddCc(emailString)
		}
	}
	if cw.Spec.Actions.Email.Bcc != "" {
		// email.AddBcc(cw.Spec.Actions.Email.Bcc)
		for _, emailString := range strings.Split(cw.Spec.Actions.Email.Bcc, ",") {
			email.AddBcc(emailString)
		}
	}
	email.SetSubject(cw.Spec.Actions.Email.Subject)

	var emailContentType = mail.TextPlain
	if cw.Spec.Actions.Email.BodyContentType == "text/html" {
		emailContentType = mail.TextHTML
	}

	email.SetBody(emailContentType, cw.Spec.Actions.Email.BodyTemplate)
	for _, f := range cw.Spec.Actions.Email.Attachments {
		email.Attach(&mail.File{FilePath: certFilesDir + "/" + f, Name: f})
		if email.Error != nil {
			return email.Error
		}
	}

	// Send email
	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}
