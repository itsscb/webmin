package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"text/template"
)

var tpl *template.Template
var em map[string]interface{}

type incident struct {
	HostName string
	UserName string
	Online   string
}

type PowerShell struct {
	powerShell string
}

func new() *PowerShell {
	ps, _ := exec.LookPath("powershell.exe")
	return &PowerShell{
		powerShell: ps,
	}
}

func (p *PowerShell) execute(args ...string) (stdOut string, stdErr string, err error) {
	args = append([]string{"-NoProfile", "-NonInteractive"}, args...)
	cmd := exec.Command(p.powerShell, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	stdOut, stdErr = stdout.String(), stderr.String()
	return
}

func pget(p *PowerShell, c string, q string, props string) map[string]interface{} {
	stdOut, stdErr, err := p.execute(c, q, "-Properties", props, "|", "ConvertTo-Json")
	if err != nil {
		log.Fatalln(err)
	}
	if stdErr != "" {
		fmt.Println(stdErr, stdOut)
	}
	resp := make(map[string]interface{})
	err = json.Unmarshal([]byte(stdOut), &resp) //"{\"key1\":0,\"key2\":0}"), &b)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

func sfbs(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "templates/assets/bootstrap/css/bootstrap.min.css")
}

func sfnv(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "templates/assets/css/Navigation-Clean.css")
}
func sfst(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "templates/assets/css/styles.css")
}
func sfjq(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "templates/assets/js/jquery.min.js")
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))

}

func main() {
	//getuser("lwescbg", new())
	http.HandleFunc("/", index)
	http.HandleFunc("/assets/bootstrap/css/bootstrap.min.css", sfbs)
	http.HandleFunc("/assets/css/Navigation-Clean.css", sfnv)
	http.HandleFunc("/assets/css/styles.css", sfst)
	http.HandleFunc("/assets/js/jquery.min.js", sfjq)
	http.HandleFunc("/user", user)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, req *http.Request) {

	posh := new()
	user := req.FormValue("username")

	/*
		rpw := req.FormValue("ResetPW")
		if rpw == "on" {
			fmt.Println("rpw", rpw)
		} else {
			fmt.Println("rpw off")
		}
		uacc := req.FormValue("UnlockAcc")
		fmt.Println(uacc)
		sprm := req.FormValue("ShowPrm")
		fmt.Println(sprm)
		exe := req.FormValue("ExtExp")
		fmt.Println(exe)
	*/

	if user == "" {
		err := tpl.ExecuteTemplate(w, "index.gohtml", em)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Fatalln(err)
		}
		return
	}
	/*
		stdOut, stdErr, err := posh.execute("Test-Connection", "-Count", "1", "-Quiet", h)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(stdErr, stdOut)
	*/
	fmt.Println("Getting Userdata of ", user, "...")
	userdata := pget(posh, "Get-ADUser", user, "DisplayName, Department, SamAccountName, Enabled, LockedOut, PasswordExpired, PasswordNeverExpires, PasswordNotRequired, CannotChangePassword, PasswordLastSet, LastBadPasswordAttempt, LastLogonDate, AccountExpirationDate, extensionAttribute13, EmailAddress, Homedirectory")
	fmt.Println("Got Userdata")
	fmt.Println(userdata)
	err := tpl.ExecuteTemplate(w, "user.gohtml", userdata) //incident{h, u, stdOut})
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}

func user(w http.ResponseWriter, req *http.Request) {
	posh := new()
	user := req.FormValue("username")

	/*
		rpw := req.FormValue("ResetPW")
		if rpw == "on" {
			fmt.Println("rpw", rpw)
		} else {
			fmt.Println("rpw off")
		}
		uacc := req.FormValue("UnlockAcc")
		fmt.Println(uacc)
		sprm := req.FormValue("ShowPrm")
		fmt.Println(sprm)
		exe := req.FormValue("ExtExp")
		fmt.Println(exe)
	*/

	if user == "" {
		err := tpl.ExecuteTemplate(w, "user.gohtml", em)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Fatalln(err)
		}
		return
	}
	/*
		stdOut, stdErr, err := posh.execute("Test-Connection", "-Count", "1", "-Quiet", h)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(stdErr, stdOut)
	*/
	fmt.Println("Getting Userdata of ", user, "...")
	userdata := pget(posh, "Get-ADUser", user, "DisplayName, Department, SamAccountName, Enabled, LockedOut, PasswordExpired, PasswordNeverExpires, PasswordNotRequired, CannotChangePassword, PasswordLastSet, LastBadPasswordAttempt, LastLogonDate, AccountExpirationDate, extensionAttribute13, EmailAddress, Homedirectory")
	fmt.Println("Got Userdata")
	fmt.Println(userdata)
	err := tpl.ExecuteTemplate(w, "user.gohtml", userdata) //incident{h, u, stdOut})
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}
