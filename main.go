package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"text/template"
	"time"
)

var tpl *template.Template

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

func pt2str(i interface{}) string {
	tmp := i.(string)[6 : len(i.(string))-5]
	ti, _ := strconv.Atoi(tmp)
	t := time.Unix(int64(ti), 0)
	return t.Format("02.01.2006 15:04")
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
		err := req.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Fatalln(err)
		}

		for key, values := range req.Form {
			fmt.Println(key, values)
			for _, value := range values {
				fmt.Println(key, value)
			}
		}
	*/
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
		em := make(map[string]interface{})
		em["username"] = ""
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
	userdata := pget(posh, "Get-ADUser", user, "DisplayName, Department, SamAccountName, Enabled, LockedOut, PasswordExpired, PasswordNeverExpires, PasswordNotRequired, CannotChangePassword, PasswordLastSet, LastBadPasswordAttempt, LastLogonDate, AccountExpirationDate, extensionAttribute13, EmailAddress, Homedirectory")
	userdata["username"] = user
	if userdata["LastLogonDate"] != nil {
		userdata["LastLogonDate"] = pt2str(userdata["LastLogonDate"])
		fmt.Println("\nLastLogonDate:", userdata["LastLogonDate"])
	} else {
		userdata["LastLogonDate"] = "None"
	}
	if userdata["PasswordLastSet"] != nil {
		userdata["PasswordLastSet"] = pt2str(userdata["PasswordLastSet"])
		fmt.Println("\nPasswordLastSet:", userdata["PasswordLastSet"])
	} else {
		userdata["PasswordLastSet"] = "None"
	}
	if userdata["LastBadPasswordAttempt"] != nil {
		userdata["LastBadPasswordAttempt"] = pt2str(userdata["LastBadPasswordAttempt"])
		fmt.Println("\nLastBadPasswordAttempt:", userdata["LastBadPasswordAttempt"])
	} else {
		userdata["LastBadPasswordAttempt"] = "None"
	}
	if userdata["AccountExpirationDate"] != nil {
		userdata["AccountExpirationDate"] = pt2str(userdata["AccountExpirationDate"])
		fmt.Println("\nAccountExpirationDate:", userdata["AccountExpirationDate"])
	} else {
		userdata["AccountExpirationDate"] = "None"
	}
	err := tpl.ExecuteTemplate(w, "user.gohtml", userdata) //incident{h, u, stdOut})
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}

func user(w http.ResponseWriter, req *http.Request) {
	posh := new()
	user := req.FormValue("username")

	if user == "" {
		em := make(map[string]interface{})
		em["username"] = ""
		err := tpl.ExecuteTemplate(w, "user.gohtml", em)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Fatalln(err)
		}
		return
	}

	fmt.Println("Getting Userdata of ", user, "...")
	userdata := pget(posh, "Get-ADUser", user, "DisplayName, Department, SamAccountName, Enabled, LockedOut, PasswordExpired, PasswordLastSet, LastBadPasswordAttempt, LastLogonDate, AccountExpirationDate, extensionAttribute13, EmailAddress, Homedirectory")
	userdata["username"] = user
	if userdata["LastLogonDate"] != nil {
		userdata["LastLogonDate"] = pt2str(userdata["LastLogonDate"])
		fmt.Println("\nLastLogonDate:", userdata["LastLogonDate"])
	} else {
		userdata["LastLogonDate"] = "None"
	}
	if userdata["PasswordLastSet"] != nil {
		userdata["PasswordLastSet"] = pt2str(userdata["PasswordLastSet"])
		fmt.Println("\nPasswordLastSet:", userdata["PasswordLastSet"])
	} else {
		userdata["PasswordLastSet"] = "None"
	}
	if userdata["LastBadPasswordAttempt"] != nil {
		userdata["LastBadPasswordAttempt"] = pt2str(userdata["LastBadPasswordAttempt"])
		fmt.Println("\nLastBadPasswordAttempt:", userdata["LastBadPasswordAttempt"])
	} else {
		userdata["LastBadPasswordAttempt"] = "None"
	}
	if userdata["AccountExpirationDate"] != nil {
		userdata["AccountExpirationDate"] = pt2str(userdata["AccountExpirationDate"])
		fmt.Println("\nAccountExpirationDate:", userdata["AccountExpirationDate"])
	} else {
		userdata["AccountExpirationDate"] = "None"
	}
	err := tpl.ExecuteTemplate(w, "user.gohtml", userdata) //incident{h, u, stdOut})
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}

