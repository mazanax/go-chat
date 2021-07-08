package db

import (
	"context"
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

func (rd *RedisDriver) CreateUser(email string, name string, encryptedPassword string) (string, error) {
	if rd.IsEmailExists(email) {
		return "", EmailAlreadyExists
	}

	// start transaction
	userUuid := uuid.NewString()
	_, err := rd.connection.TxPipelined(rd.ctx, func(pipe redis.Pipeliner) error {
		_, err := pipe.HSet(
			rd.ctx,
			fmt.Sprintf("user(%s)", userUuid),
			map[string]interface{}{
				"id":        userUuid,
				"email":     strings.ToLower(email),
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
	val, err := rd.connection.HGetAll(rd.ctx, fmt.Sprintf("user(%s)", id)).Result()
	switch {
	case err == redis.Nil || len(val) == 0:
		return models.User{}, UserNotFound
	case err != nil:
		logger.Fatal("Redis connection failed: %s", err.Error())
	}

	createdAt, _ := strconv.Atoi(val["createdAt"])
	updatedAt, _ := strconv.Atoi(val["updatedAt"])
	return models.User{
		ID:        val["id"],
		Email:     val["email"],
		Name:      val["name"],
		Password:  val["password"],
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
