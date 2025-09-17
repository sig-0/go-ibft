package message

type Cache[M message] struct {
	filterFn func(M) bool
	seen     map[string]struct{}
	messages []M
}

func NewCache[M message](filterFn func(M) bool) *Cache[M] {
	return &Cache[M]{
		filterFn: filterFn,
		messages: make([]M, 0),
		seen:     make(map[string]struct{}),
	}
}

func (c *Cache[M]) Add(messages ...M) {
	for _, msg := range messages {
		sender := string(msg.GetSender())
		if _, ok := c.seen[sender]; ok {
			continue
		}

		c.seen[sender] = struct{}{}

		if !c.filterFn(msg) {
			continue
		}

		c.messages = append(c.messages, msg)
	}
}

func (c *Cache[M]) Get() []M {
	return c.messages
}
