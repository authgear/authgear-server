package pubsub

type Publisher struct {
	RedisPool RedisPool
}

func (p *Publisher) Publish(channelName string, data []byte) error {
	redisConn := p.RedisPool.Get()
	defer redisConn.Close()
	_, err := redisConn.Do("PUBLISH", channelName, data)
	return err
}
