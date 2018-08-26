package backend

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type Client struct {
	Client *redis.Client
}

func NewBackend(host, port string) (*Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     "",
		DB:           0,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})

	err := c.Ping().Err()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to redis at %s:%s", host, port)
	}

	return &Client{
		Client: c,
	}, nil
}

func (b *Client) SetTimeId(id, val string) error {
	err := b.Client.Set(id, val, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (b *Client) GetTimeId(id string) (string, error) {
	val, err := b.Client.Get(id).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (b *Client) DeleteTimeId(id string) error {
	val, err := b.Client.Del(id).Result()
	if err != nil {
		return err
	}
	if val == 0 {
		return redis.Nil
	}
	return nil
}

func (b *Client) NotFoundErrCheck(err error) bool { return err == redis.Nil }
