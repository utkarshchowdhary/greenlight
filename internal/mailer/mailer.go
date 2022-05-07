package mailer

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Below we declare a new variable with the type embed.FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.
// ↓↓↓
//go:embed "templates"
var templateFS embed.FS

// Define a Mailer struct which contains a sendgrid.Client instance
// and the sender information for your emails (the name and address you
// want the email to be from).
type Mailer struct {
	client *sendgrid.Client
	sender string
}

func New(apikey, sender string) Mailer {
	// Initialize a new sendgrid.Client instance with the given apikey.
	client := sendgrid.NewSendClient(apikey)

	// Return a Mailer instance containing the client and sender information.
	return Mailer{
		client: client,
		sender: sender,
	}
}

// Define a Send() method on the Mailer type. This takes the recipient email address
// as the first parameter, the name of the file containing the templates, and any
// dynamic data for the templates as an interface{} parameter.
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// Use the ParseFS() method to parse the required template file from the embedded
	// file system.
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Execute the "plainBody" template and store the result in the plainBody variable.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Execute the "htmlBody" template and store the result in the htmlBody variable.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewSingleEmail() function to construct the message.
	msg := mail.NewSingleEmail(
		mail.NewEmail("", m.sender),
		subject.String(),
		mail.NewEmail("", recipient),
		plainBody.String(),
		htmlBody.String(),
	)

	// Try sending the email up to three times before aborting and returning the final
	// error. We sleep for 500 milliseconds between each attempt.
	for i := 1; i <= 3; i++ {
		// Create a context with a 5-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call the SendWithContext() method on the client, passing in the message to send.
		_, err = m.client.SendWithContext(ctx, msg)
		if err == nil {
			return nil
		}

		// If it didn't work, sleep for a short time and retry.
		time.Sleep(500 * time.Millisecond)
	}

	return err
}
