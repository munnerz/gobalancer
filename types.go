package main

type LoadBalancer interface {
	Run(chan error) error
	Stop() error
}
