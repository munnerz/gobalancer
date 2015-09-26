package main

type LoadBalancer interface {
	Run() error
	Stop() error
}
