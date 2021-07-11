package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mazanax/go-chat/app/logger"
	"github.com/mazanax/go-chat/app/models"
	"strconv"
	"strings"
	"time"
)

type RedisDriver struct {
	ctx        context.Context
	connection *redis.Client
}

func NewRedisDriver(ctx context.Context, addr string, password string, defaultDb int) RedisDriver {
	return RedisDriver{
		ctx: ctx,
		connection: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       defaultDb,
		}),
	}
}

// region UserRepository

func (rd *RedisDriver) IsEmailExists(email string) bool {
	val, err := rd.connection.HExists(rd.ctx, "emails", strings.ToLower(email)).Result()
	switch {
	case err == redis.Nil:
		return false
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	return val
}

func (rd *RedisDriver) IsUsernameExists(username string) bool {
	val, err := rd.connection.HExists(rd.ctx, "usernames", strings.ToLower(username)).Result()
	switch {
	case err == redis.Nil:
		return false
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	return val
}

func (rd *RedisDriver) CreateUser(email string, username string, name string, encryptedPassword string) (string, error) {
	if rd.IsEmailExists(email) {
		return "", EmailAlreadyExists
	}

	if rd.IsUsernameExists(username) {
		return "", UsernameAlreadyExists
	}

	// start transaction
	userUuid := uuid.NewString()
	_, err := rd.connection.TxPipelined(rd.ctx, func(pipe redis.Pipeliner) error {
		_, err := pipe.HSet(
			rd.ctx,
			fmt.Sprintf("user:%s", userUuid),
			map[string]interface{}{
				"id":        userUuid,
				"email":     strings.ToLower(email),
				"username":  username,
				"name":      name,
				"password":  encryptedPassword,
				"createdAt": time.Now().Unix(),
				"updatedAt": time.Now().Unix(),
			},
		).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		_, err = pipe.HSet(rd.ctx, "emails", strings.ToLower(email), userUuid).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		_, err = pipe.HSet(rd.ctx, "usernames", strings.ToLower(username), userUuid).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		_, err = pipe.SAdd(rd.ctx, "users", userUuid).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		return nil
	})

	if err != nil {
		return "", UserNotCreated
	}

	return userUuid, nil
}

func (rd *RedisDriver) GetUser(id string) (models.User, error) {
	val, err := rd.connection.HGetAll(rd.ctx, fmt.Sprintf("user:%s", id)).Result()
	switch {
	case errors.Is(err, redis.Nil) || len(val) == 0:
		return models.User{}, UserNotFound
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	createdAt, _ := strconv.Atoi(val["createdAt"])
	updatedAt, _ := strconv.Atoi(val["updatedAt"])
	return models.User{
		ID:        val["id"],
		Email:     val["email"],
		Username:  val["username"],
		Name:      val["name"],
		Password:  val["password"],
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (rd *RedisDriver) GetUsers() []models.User {
	var cursor uint64 = 0
	users, cursor, err := rd.connection.SScan(rd.ctx, "users", cursor, "", 0).Result()
	if err != nil {
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	var result []models.User
	for _, user := range users {
		model, err := rd.GetUser(user)
		if err != nil {
			logger.Error("[GetUsers] Cannot get user %s\n", err)
			continue
		}

		result = append(result, model)
	}

	return result
}

func (rd *RedisDriver) FindUserByEmail(email string) (models.User, error) {
	if !rd.IsEmailExists(email) {
		return models.User{}, UserNotFound
	}

	userId, err := rd.connection.HGet(rd.ctx, "emails", strings.ToLower(email)).Result()
	switch {
	case errors.Is(err, redis.Nil) || len(userId) == 0:
		return models.User{}, UserNotFound
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	return rd.GetUser(userId)
}

// endregion

// region TokenRepository

func (rd *RedisDriver) CreateToken(user *models.User, randomString string, duration time.Duration) (string, error) {
	// start transaction
	tokenUuid := uuid.NewString()
	_, err := rd.connection.TxPipelined(rd.ctx, func(pipe redis.Pipeliner) error {
		_, err := pipe.HSet(
			rd.ctx,
			fmt.Sprintf("token:%s", tokenUuid),
			map[string]interface{}{
				"id":        tokenUuid,
				"userId":    user.ID,
				"token":     randomString,
				"createdAt": time.Now().Unix(),
				"expireAt":  time.Now().Add(duration).Unix(),
			},
		).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}
		_, err = pipe.Expire(rd.ctx, fmt.Sprintf("token:%s", tokenUuid), duration).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}
		_, err = pipe.Set(rd.ctx, fmt.Sprintf("token_to_uuid:%s", randomString), tokenUuid, duration).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		_, err = pipe.SAdd(rd.ctx, fmt.Sprintf("user_tokens:%s", user.ID), tokenUuid).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		return nil
	})

	if err != nil {
		return "", TokenNotCreated
	}

	return tokenUuid, nil
}

func (rd *RedisDriver) GetToken(id string) (models.AccessToken, error) {
	val, err := rd.connection.HGetAll(rd.ctx, fmt.Sprintf("token:%s", id)).Result()
	switch {
	case errors.Is(err, redis.Nil) || len(val) == 0:
		return models.AccessToken{}, TokenNotFound
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	createdAt, _ := strconv.Atoi(val["createdAt"])
	expireAt, _ := strconv.Atoi(val["expireAt"])
	return models.AccessToken{
		ID:        val["id"],
		UserID:    val["userId"],
		Token:     val["token"],
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
	}, nil
}

func (rd *RedisDriver) FindTokenByString(token string) (models.AccessToken, error) {
	tokenUUID, err := rd.connection.Get(rd.ctx, fmt.Sprintf("token_to_uuid:%s", token)).Result()
	switch {
	case errors.Is(err, redis.Nil) || len(tokenUUID) == 0:
		return models.AccessToken{}, TokenNotFound
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	return rd.GetToken(tokenUUID)
}

func (rd *RedisDriver) RemoveToken(token models.AccessToken) error {
	_, err := rd.connection.TxPipelined(rd.ctx, func(pipe redis.Pipeliner) error {
		_, err := pipe.Del(rd.ctx, fmt.Sprintf("token_to_uuid:%s", token.Token)).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		_, err = pipe.Del(rd.ctx, fmt.Sprintf("token:%s", token.ID)).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		_, err = pipe.SRem(rd.ctx, fmt.Sprintf("user_tokens:%s", token.UserID), token.ID).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		return nil
	})

	return err
}

// endregion

// region TicketRepository

func (rd *RedisDriver) CreateTicket(accessToken *models.AccessToken, randomString string, duration time.Duration) error {
	_, err := rd.connection.TxPipelined(rd.ctx, func(pipe redis.Pipeliner) error {
		_, err := pipe.HSet(
			rd.ctx,
			fmt.Sprintf("ticket:%s", randomString),
			map[string]interface{}{
				"userId":    accessToken.UserID,
				"tokenId":   accessToken.ID,
				"ticket":    randomString,
				"createdAt": time.Now().Unix(),
				"expireAt":  time.Now().Add(duration).Unix(),
			},
		).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}
		_, err = pipe.Expire(rd.ctx, fmt.Sprintf("ticket:%s", randomString), duration).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		return nil
	})

	return err
}

func (rd *RedisDriver) GetTicket(ticket string) (models.Ticket, error) {
	val, err := rd.connection.HGetAll(rd.ctx, fmt.Sprintf("ticket:%s", ticket)).Result()
	switch {
	case errors.Is(err, redis.Nil) || len(val) == 0:
		return models.Ticket{}, TicketNotFound
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	createdAt, _ := strconv.Atoi(val["createdAt"])
	expireAt, _ := strconv.Atoi(val["expireAt"])
	return models.Ticket{
		UserID:    val["userId"],
		TokenID:   val["tokenId"],
		Ticket:    val["ticket"],
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
	}, nil
}

func (rd *RedisDriver) RemoveTicket(ticket models.Ticket) error {
	_, err := rd.connection.Del(rd.ctx, fmt.Sprintf("ticket:%s", ticket.Ticket)).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil
	}

	return err
}

// endregion

// region OnlineRepository

func (rd *RedisDriver) GetOnlineUsers() []string {
	var cursor uint64 = 0
	users, cursor, err := rd.connection.SScan(rd.ctx, "online", cursor, "", 0).Result()
	if err != nil {
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	return users
}

func (rd *RedisDriver) CreateUserOnline(userUUID string) error {
	_, err := rd.connection.SAdd(rd.ctx, "online", userUUID).Result()
	if err != nil {
		return err
	}

	return nil
}

func (rd *RedisDriver) RemoveUserOnline(userUUID string) error {
	_, err := rd.connection.SRem(rd.ctx, "online", userUUID).Result()
	if err != nil {
		return err
	}

	return nil
}

// endregion

// region MessageRepository

func (rd *RedisDriver) StoreMessage(userID string, messageType int, messageUUID string, text string) (string, error) {
	_, err := rd.connection.TxPipelined(rd.ctx, func(pipe redis.Pipeliner) error {
		_, err := pipe.HSet(
			rd.ctx,
			fmt.Sprintf("message:%s", messageUUID),
			map[string]interface{}{
				"id":        messageUUID,
				"userId":    userID,
				"createdAt": time.Now().Unix(),
				"type":      messageType,
				"text":      text,
			},
		).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		_, err = pipe.LPush(rd.ctx, "messages", messageUUID).Result()
		if err != nil {
			_ = pipe.Discard()
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return messageUUID, nil
}

func (rd *RedisDriver) GetMessage(messageUUID string) (models.Message, error) {
	val, err := rd.connection.HGetAll(rd.ctx, fmt.Sprintf("message:%s", messageUUID)).Result()
	switch {
	case errors.Is(err, redis.Nil) || len(val) == 0:
		return models.Message{}, MessageNotFound
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	createdAt, _ := strconv.Atoi(val["createdAt"])
	messageType, _ := strconv.Atoi(val["type"])
	return models.Message{
		ID:        val["id"],
		UserID:    val["userId"],
		CreatedAt: createdAt,
		Type:      messageType,
		Text:      val["text"],
	}, nil
}

func (rd *RedisDriver) GetMessages(limit int) []models.Message {
	messages, err := rd.connection.LRange(rd.ctx, "messages", 0, int64(limit)).Result()
	if err != nil {
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	var result []models.Message
	for _, message := range messages {
		model, err := rd.GetMessage(message)
		if err != nil {
			logger.Error("[GetMessages] Cannot get message %s\n", err)
			continue
		}

		result = append(result, model)
	}

	return result
}

// endregion
