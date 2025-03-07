package main

import (
	"errors"
	"fmt"
	"log"
)

// TransactionStatus represents the current state of a transaction
type TransactionStatus string

const (
	Pending     TransactionStatus = "PENDING"
	Committed   TransactionStatus = "COMMITTED"
	Compensated TransactionStatus = "COMPENSATED"
	Failed      TransactionStatus = "FAILED"
)

// Transaction represents a single step in a Saga
type Transaction struct {
	Name             string
	Action           func() error
	CompensateAction func() error
	Status           TransactionStatus
}

// Saga orchestrates a series of distributed transactions
type Saga struct {
	Transactions   []*Transaction
	CompletedSteps []*Transaction
}

// NewSaga creates a new Saga orchestrator
func NewSaga(transactions []*Transaction) *Saga {
	return &Saga{
		Transactions:   transactions,
		CompletedSteps: []*Transaction{},
	}
}

// Execute runs the entire saga transaction sequence
func (s *Saga) Execute() error {
	for _, transaction := range s.Transactions {
		// Execute the forward action
		err := transaction.Action()
		if err != nil {
			log.Printf("Transaction %s failed: %v", transaction.Name, err)
			// If any transaction fails, start compensation
			s.Compensate()
			return errors.New("saga transaction failed")
		}

		// Mark transaction as committed and add to completed steps
		transaction.Status = Committed
		s.CompletedSteps = append(s.CompletedSteps, transaction)
	}

	return nil
}

// Compensate rolls back completed transactions in reverse order
func (s *Saga) Compensate() {
	log.Println("Starting compensation...")

	// Iterate through completed steps in reverse
	for i := len(s.CompletedSteps) - 1; i >= 0; i-- {
		transaction := s.CompletedSteps[i]

		// Execute compensating action
		err := transaction.CompensateAction()
		if err != nil {
			log.Printf("Compensation failed for %s: %v", transaction.Name, err)
			transaction.Status = Failed
		} else {
			transaction.Status = Compensated
		}
	}
}

// Example usage with an order processing scenario
func main() {
	// Create order transactions
	transactions := []*Transaction{
		{
			Name: "Create Order",
			Action: func() error {
				fmt.Println("Creating order...")
				return nil
			},
			CompensateAction: func() error {
				fmt.Println("Cancelling order...")
				return nil
			},
		},
		{
			Name: "Reserve Inventory",
			Action: func() error {
				fmt.Println("Reserving inventory...")
				return nil
			},
			CompensateAction: func() error {
				fmt.Println("Releasing inventory...")
				return nil
			},
		},
		{
			Name: "Process Payment",
			Action: func() error {
				fmt.Println("Processing payment...")
				// Simulate a payment failure
				return errors.New("payment processing failed")
			},
			CompensateAction: func() error {
				fmt.Println("Refunding payment...")
				return nil
			},
		},
	}

	// Create and execute Saga
	saga := NewSaga(transactions)

	if err := saga.Execute(); err != nil {
		fmt.Println("Saga transaction failed and compensated")
	} else {
		fmt.Println("Saga transaction completed successfully")
	}
}
