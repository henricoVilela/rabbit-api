package rabbit

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Notification struct {
	UserId      string `json:"userId"`
	Application string `json:"application"`
	Message     string `json:"message"`
}

var EXCHANGE_NAME = "apps_notify"
var connRb *amqp.Connection

var (
	timer              *time.Timer
	inactivityDuration = 5 * time.Minute // Defina a duração de inatividade desejada
)

func SendMessage(n Notification) error {

	if err := useConnection(); err != nil {
		return err
	}

	ch, err := connRb.Channel()
	if err != nil {
		log.Printf("falhou ao buscar um canal no rabbit: %v", err)
		return err
	}

	defer ch.Close()

	initExchange(ch)
	initQueueAndBind(ch, n)

	body, err := json.Marshal(n)
	if err != nil {
		return errors.New("json malformado")
	}

	key := strings.ToLower(n.Application) + ".user." + strings.ToLower(n.UserId)

	err = ch.Publish(
		EXCHANGE_NAME, // exchange
		key,           // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})

	if err != nil {
		msg := "falhou ao enviar a mensagem ao rabbit"
		log.Printf("%s: %v", msg, err)
		return errors.New(msg)
	}

	log.Printf(" [x] enviado %s", body)

	return nil
}

func initExchange(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		EXCHANGE_NAME, // nome da exchange
		"topic",       // tipo
		true,          // durável
		false,         // auto-delete
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)

	return err
}

func initQueueAndBind(ch *amqp.Channel, n Notification) error {

	queueName := strings.ToLower(n.Application) + ".user." + strings.ToLower(n.UserId)
	_, err := ch.QueueDeclare(
		queueName, // nome da fila
		true,      // durável
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		log.Printf("falhou ao criar fila no rabbit: %v", err)
		return fmt.Errorf("falhou ao criar a fila %s no rabbitmq", queueName)
	}

	err = ch.QueueBind(
		queueName,     // nome da fila
		queueName,     // chave de roteamento
		EXCHANGE_NAME, // nome da exchange
		false,
		nil,
	)

	if err != nil {
		msg := "falhou ao fazer o bind na exchange"
		log.Printf("%s: %v", msg, err)
		return fmt.Errorf("%s", msg)
	}

	return nil
}

func connect() error {
	var err error

	if connRb == nil || (connRb != nil && connRb.IsClosed()) {

		connRb, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			log.Printf("falhou ao conectar ao rabbitmq: %v", err)
			return errors.New("falhou ao conectar ao rabbitmq")
		}
	}

	log.Println("Conectado ao Rabbit")
	return nil
}

func Disconnect() {
	if connRb != nil {
		connRb.Close()
		connRb = nil

		log.Println("Conexão foi fechada")
	}
}

func resetTimer() {
	if timer != nil {
		timer.Stop()
	}

	timer = time.AfterFunc(inactivityDuration, func() {
		Disconnect()
	})
}

func useConnection() error {
	resetTimer()
	if connRb == nil {
		return connect() // Substitua pela lógica para criar sua conexão
	}
	return nil
}
