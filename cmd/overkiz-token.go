package main

import (
	"flag"
	"fmt"
	"os"
	"overkiz-adapter/internal/domain"
)

func main() {
	region := new(string)
	username := new(string)
	password := new(string)
	pod := new(string)
	label := new(string)
	uuid := new(string)

	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	loginCmd.StringVar(region, "region", "", "Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	loginCmd.StringVar(username, "username", "", "Username")
	loginCmd.StringVar(password, "password", "", "Password")

	logoutCmd := flag.NewFlagSet("logout", flag.ExitOnError)
	logoutCmd.StringVar(region, "region", "", "Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listCmd.StringVar(region, "region", "", "Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	listCmd.StringVar(pod, "pin", "", "The PIN of the gateway")

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createCmd.StringVar(region, "region", "", "Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	createCmd.StringVar(pod, "pin", "", "The PIN of the gateway")
	createCmd.StringVar(label, "label", "Machnos overkiz-token", "The label of the token")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteCmd.StringVar(region, "region", "", "Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	deleteCmd.StringVar(pod, "pin", "", "The PIN of the gateway")
	deleteCmd.StringVar(uuid, "uuid", "", "The uuid of the token")

	if len(os.Args) < 2 {
		printTokenUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "login":
		err := loginCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if *region == "" || *username == "" || *password == "" {
			printLoginUsage()
			os.Exit(1)
		}
		api, err := domain.NewOverkizTokenApi(*region)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = api.Login(*username, *password)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "logout":
		err := logoutCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if *region == "" {
			printLogoutUsage()
			os.Exit(1)
		}
		api, err := domain.NewOverkizTokenApi(*region)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = api.Logout()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "list":
		err := listCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if *region == "" || *pod == "" {
			printTokenListUsage()
			os.Exit(1)
		}
		api, err := domain.NewOverkizTokenApi(*region)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = api.PrintTokens(*pod)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "create":
		err := createCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if *region == "" || *pod == "" {
			printTokenCreateUsage()
			os.Exit(1)
		}
		api, err := domain.NewOverkizTokenApi(*region)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = api.CreateAndPrintToken(*pod, *label)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "delete":
		err := deleteCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if *region == "" || *pod == "" || *uuid == "" {
			printTokenDeleteUsage()
			os.Exit(1)
		}
		api, err := domain.NewOverkizTokenApi(*region)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = api.DeleteToken(*pod, *uuid)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		printTokenUsage()
		os.Exit(1)
	}
}

func printTokenUsage() {
	println("Manages Overkiz tokens")
	println("")
	println("Usage:")
	println("  overkiz-token [command] [options]")
	println("")
	println("Available Commands:")
	println("  login      Login to Overkiz")
	println("  logout     Logout from Overkiz")
	println("  list       List all tokens")
	println("  create     Create a new token")
	println("  delete     Delete an existing token")
}

func printLoginUsage() {
	println("Login to Overkiz")
	println("")
	println("Usage:")
	println("  overkiz-token login [options]")
	println("")
	println("Required Options:")
	println("  --region   Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	println("  --username Your Overkiz username")
	println("  --password Your Overkiz password")
}

func printLogoutUsage() {
	println("Logout from Overkiz")
	println("")
	println("Usage:")
	println("  overkiz-token logout [options]")
	println("")
	println("Required Options:")
	println("  --region   Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
}

func printTokenListUsage() {
	println("List Overkiz tokens")
	println("")
	println("Usage:")
	println("  overkiz-token list [options]")
	println("")
	println("Required Options:")
	println("  --region   Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	println("  --pin      The PIN of the gateway")
}

func printTokenCreateUsage() {
	println("Create an Overkiz token")
	println("")
	println("Usage:")
	println("  overkiz-token create [options]")
	println("")
	println("Required Options:")
	println("  --region   Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	println("  --pin      The PIN of the gateway")
	println("Optional Options:")
	println("  --label    The label of the new token")
}

func printTokenDeleteUsage() {
	println("Delete an Overkiz token")
	println("")
	println("Usage:")
	println("  overkiz-token delete [options]")
	println("")
	println("Required Options:")
	println("  --region   Region, one of \"europe\", \"middle east\", \"africa\", \"asia\", \"pacific\" or \"north america\"")
	println("  --pin      The PIN of the gateway")
	println("  --uuid     The UUID of the token that should be deleted")
}
