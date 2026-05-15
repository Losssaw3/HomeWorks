package mail

import (
	"context"
)

type Mailer struct {
}

func (m *Mailer) SendConfirmation(ctx context.Context, email string, body string) error {
	//Заглушка
	return nil
}
