package main

import "fmt"

type PaymentEvent struct {
	ID          string `json:"id"`
	CustomerID  string `json:"customer_id"`
	AmountCents int64  `json:"amount_cents"`
	Country     string `json:"country"`
	Attempts    int    `json:"attempts"`
}

type Decision string

const (
	Approve Decision = "approve"
	Review  Decision = "review"
	Decline Decision = "decline"
)

func assess(e PaymentEvent) Decision {
	if e.AmountCents <= 0 || e.CustomerID == "" {
		return Decline
	}
	if e.Attempts >= 3 || e.AmountCents >= 500000 {
		return Review
	}
	return Approve
}

func HandlePayment(c *InfraiClient, e PaymentEvent) (Decision, error) {
	d := assess(e)
	if d == Decline {
		err := c.Capture(map[string]any{
			"title":       "payment declined by validation",
			"message":     fmt.Sprintf("payment %s failed validation", e.ID),
			"level":       "warning",
			"fingerprint": []string{"payment-validation", e.Country},
			"exception":   "invalid payment event",
			"context":     e,
		})
		if err != nil {
			return d, err
		}
	}
	return d, nil
}

func main() {
	c, err := NewInfraiClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	d, err := HandlePayment(c, PaymentEvent{ID: "pay_demo", CustomerID: "cust_42", AmountCents: 0, Country: "US"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("payment decision: %s\n", d)
}
