package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"gopkg.in/gomail.v2"
)

// Configurações
const (
	websiteURL    = "https://exemplo.com/api/data"
	dynamoDBTable = "NomeDaTabelaNoDynamoDB"
	awsRegion     = "us-east-1" // Altere para a região desejada
)

// Estrutura para o corpo JSON
type Data struct {
	// Defina a estrutura do JSON conforme necessário
}

func main() {
	// Obter dados do site
	data, err := fetchData()
	if err != nil {
		log.Fatal("Erro ao obter dados do site:", err)
	}

	// Enviar e-mail apenas se a obtenção de dados for bem-sucedida
	err = sendEmail(data)
	if err != nil {
		log.Fatal("Erro ao enviar e-mail:", err)
	}

	// Registrar dados no DynamoDB
	err = logToDynamoDB(data)
	if err != nil {
		log.Fatal("Erro ao registrar dados no DynamoDB:", err)
	}

	fmt.Println("Operação concluída com sucesso!")
}

func fetchData() (*Data, error) {
	resp, err := http.Get(websiteURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Resposta do site não foi bem-sucedida. Código: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func sendEmail(data *Data) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "seuemail@gmail.com")
	m.SetHeader("To", "destinatario@gmail.com")
	m.SetHeader("Subject", "Dados do Site")
	m.SetBody("text/html", fmt.Sprintf("<p>%+v</p>", data))

	d := gomail.NewDialer("smtp.gmail.com", 587, "seuemail@gmail.com", "suasenha")

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func logToDynamoDB(data *Data) error {
	// Configurar sessão AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		return err
	}

	// Criar serviço DynamoDB
	svc := dynamodb.New(sess)

	// Converter struct para mapa
	item, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return err
	}

	// Adicionar timestamp
	item["Timestamp"] = aws.String(time.Now().Format(time.RFC3339))

	// Configurar input para PutItem
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(dynamoDBTable),
	}

	// Executar operação PutItem
	_, err = svc.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}
