package app

import "github.com/mazanax/go-chat/app/models"

func mapUserToJson(user models.User, withEmail bool) models.JsonUser {
	email := user.Email
	if !withEmail {
		email = ""
	}

	return models.JsonUser{
		ID:        user.ID,
		Name:      user.Name,
		Email:     email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func mapAccessTokenToJson(token models.AccessToken) models.JsonAccessToken {
	return models.JsonAccessToken{
		Token:     token.Token,
		CreatedAt: token.CreatedAt,
		ExpireAt:  token.ExpireAt,
	}
}

func mapTicketToJson(ticket models.Ticket) models.JsonTicket {
	return models.JsonTicket{
		Ticket:    ticket.Ticket,
		CreatedAt: ticket.CreatedAt,
		ExpireAt:  ticket.ExpireAt,
	}
}

func mapMessageToJson(message models.Message) models.JsonMessage {
	return models.JsonMessage{
		ID:        message.ID,
		UserID:    message.UserID,
		Type:      message.Type,
		CreatedAt: message.CreatedAt,
		Text:      message.Text,
	}
}
