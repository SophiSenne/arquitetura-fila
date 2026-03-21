package messaging

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type channelEntry struct {
	ch     *amqp.Channel
	broken bool
}

type Pool struct {
	mu       sync.Mutex
	conn     *amqp.Connection
	url      string
	channels chan *channelEntry
	size     int
}

var (
	globalPool *Pool
	once       sync.Once
)

func GetPool() (*Pool, error) {
	var initErr error
	once.Do(func() {
		user := os.Getenv("RABBITMQ_USER")
		password := os.Getenv("RABBITMQ_PASSWORD")
		host := os.Getenv("RABBITMQ_HOST")
		port := os.Getenv("RABBITMQ_PORT")

		if user == "" || password == "" || host == "" || port == "" {
			initErr = fmt.Errorf(
				"variáveis de ambiente do RabbitMQ não definidas (USER=%q, HOST=%q, PORT=%q, PASSWORD vazia=%v)",
				user, host, port, password == "",
			)
			return
		}

		url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
		p, err := newPool(url, 10) 
		if err != nil {
			initErr = err
			return
		}
		globalPool = p
	})

	if initErr != nil {
		return nil, initErr
	}
	return globalPool, nil
}

func newPool(url string, size int) (*Pool, error) {
	p := &Pool{
		url:      url,
		size:     size,
		channels: make(chan *channelEntry, size),
	}

	if err := p.dial(); err != nil {
		return nil, err
	}

	for i := 0; i < size; i++ {
		entry, err := p.newEntry()
		if err != nil {
			return nil, fmt.Errorf("erro ao criar channel %d/%d: %w", i+1, size, err)
		}
		p.channels <- entry
	}

	log.Printf("[Pool] Inicializado: %d channels prontos", size)
	return p, nil
}

func (p *Pool) Acquire() (*channelEntry, error) {
	entry := <-p.channels 

	if entry.broken || entry.ch.IsClosed() {
		log.Printf("[Pool] Channel inválido detectado, recriando...")
		if err := p.healEntry(entry); err != nil {
			p.channels <- entry
			return nil, fmt.Errorf("não foi possível recriar channel: %w", err)
		}
	}

	return entry, nil
}

func (p *Pool) Release(entry *channelEntry) {
	p.channels <- entry
}

func (p *Pool) healEntry(entry *channelEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn == nil || p.conn.IsClosed() {
		log.Printf("[Pool] Conexão perdida, reconectando...")
		if err := p.dial(); err != nil {
			return err
		}
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("erro ao abrir channel: %w", err)
	}

	entry.ch = ch
	entry.broken = false
	return nil
}

func (p *Pool) newEntry() (*channelEntry, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, err
	}
	return &channelEntry{ch: ch}, nil
}

func (p *Pool) dial() error {
	backoff := 1 * time.Second
	for i := 0; i < 10; i++ {
		conn, err := amqp.Dial(p.url)
		if err == nil {
			p.conn = conn
			log.Printf("[Pool] Conexão AMQP estabelecida")
			return nil
		}
		log.Printf("[Pool] RabbitMQ indisponível (%v), tentando novamente em %s...", err, backoff)
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
	return fmt.Errorf("não foi possível conectar ao RabbitMQ após 10 tentativas")
}

func (p *Pool) Close() {
	close(p.channels)
	for entry := range p.channels {
		_ = entry.ch.Close()
	}
	if p.conn != nil && !p.conn.IsClosed() {
		_ = p.conn.Close()
	}
	log.Printf("[Pool] Pool encerrado")
}