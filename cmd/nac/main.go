package main

import "fmt"

type AuthSecret struct{
	name string
	keyUser string
	keyPass string
}

type Apply struct{
	repoUrl string
	authSecret AuthSecret
}

type Helm struct{
	repoUrl string
	path string
	targetRevision string
	chart string
	name string
}



func main(){
	fmt. Println("hello")
}

