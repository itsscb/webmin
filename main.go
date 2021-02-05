package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
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

func pget(p *PowerShell, c string) string {
	stdOut, stdErr, err := p.execute(c)
	if err != nil {
		log.Fatalln(err)
	}
	if stdErr != "" {
		fmt.Println(stdErr, stdOut)
	}
	return stdOut
}

func p2j(s string) map[string]interface{} {
	resp := make(map[string]interface{})
	err := json.Unmarshal([]byte(s), &resp)
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
	tf := false
	posh := new()
	user := req.FormValue("username")

	if user == "" {
		em := make(map[string]interface{})
		em["username"] = ""
		em["hostname"] = ""
		err := tpl.ExecuteTemplate(w, "index.gohtml", em)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Fatalln(err)
		}
		return
	}

	ud := pget(posh, "Get-ADUser "+user+" -Properties DisplayName, Department, SamAccountName, Enabled, LockedOut, PasswordExpired, PasswordNeverExpires, PasswordNotRequired, CannotChangePassword, PasswordLastSet, LastBadPasswordAttempt, LastLogonDate, AccountExpirationDate, extensionAttribute13, EmailAddress, Homedirectory | ConvertTo-Json")
	userdata := p2j(ud)
	userdata["username"] = user
	userdata["vpnuser"] = pget(posh, "$t = Get-ADPrincipalGroupMembership "+user+" | Select-Object -Property Name | Where-Object { $_.Name -eq '.LWE-VPN-LWE_User' -or $_.Name -eq '.LWE-VPN-LWE_Admin' }; $t -ne $null")
	if userdata["LastLogonDate"] != nil {
		userdata["LastLogonDate"] = pt2str(userdata["LastLogonDate"])
	} else {
		userdata["LastLogonDate"] = "None"
	}
	if userdata["PasswordLastSet"] != nil {
		userdata["PasswordLastSet"] = pt2str(userdata["PasswordLastSet"])
	} else {
		userdata["PasswordLastSet"] = "None"
	}
	if userdata["LastBadPasswordAttempt"] != nil {
		userdata["LastBadPasswordAttempt"] = pt2str(userdata["LastBadPasswordAttempt"])
	} else {
		userdata["LastBadPasswordAttempt"] = "None"
	}
	if userdata["AccountExpirationDate"] != nil {
		userdata["AccountExpirationDate"] = pt2str(userdata["AccountExpirationDate"])
	} else {
		userdata["AccountExpirationDate"] = "None"
	}

	uf := req.FormValue("UserForm")

	switch uf {
	case "ResetPW":
		fmt.Printf("ResetPW")
		userdata["response"] = pget(posh, "Set-ADAccountPassword -Reset -NewPassword (ConvertTo-SecureString -String 'Liebherr1!' -AsPlainText -Force) -Identity "+user)
		userdata["command"] = "Reset Password"
		tf = true
	case "UnlockAcc":
		fmt.Printf("UnlockAcc")
		userdata["response"] = pget(posh, "Unlock-ADAccount -Identity "+user+"; $?")
		userdata["command"] = "Unlock Account"
		tf = true
	case "ShowPerm":
		fmt.Printf("ShowPerm")
		groups := pget(posh, "Get-ADPrincipalGroupMembership "+user+" | Select-Object -Property Name | Sort-Object -Property Name #| ConvertTo-Json")
		groups = strings.Replace(groups, "\n", "<br>", -1)
		fmt.Printf(groups)
		userdata["response"] = groups
		userdata["command"] = "Show Permissions"
	case "ExtExp":
		fmt.Printf("ExtExp")
		userdata["response"] = pget(posh, "Set-ADUser "+user+" -AccountExpirationDate (Get-Date).AddDays(14); $?")
		userdata["command"] = "Extent Account Expirationdate"
		tf = true
	case "LastHost":
		fmt.Printf("LastHost")
		userdata["response"] = pget(posh, "Unlock-ADAccount -Identity "+user+"; $?")
		userdata["command"] = "Get Last used Computer"
	case "EnableJabber":
		fmt.Printf("Enable Jabber")
		userdata["response"] = exec.Command("\\\\lwesv0170\\itsm_tools\\jabber\\jabber.exe", user)
		userdata["command"] = "Enable Cisco Jabber"
		tf = true
	case "AddVPN":
		fmt.Printf("AddVPN")
		userdata["response"] = pget(posh, "$u = Get-ADUser "+user+" -Properties * ; $g = Get-ADGroup '.LWE-VPN-LWE_User' ; Set-ADObject -Identity $g -Add @{member=$u.DistinguishedName}; $?")
		userdata["command"] = "Add User to VPN-Group"
		tf = true
	case "EmergencyVPN":
		fmt.Printf("EmergencyVPN")
		userdata["response"] = pget(posh, "Unlock-ADAccount -Identity "+user)
		userdata["command"] = "Assign Emergency-VPN"
	}
	if tf == true {
		userdata["response"] = userdata["response"].(string)[:4]
		if userdata["response"].(string) == "True" {
			userdata["response"] = "Success"
		} else {
			userdata["response"] = "Failed"
		}
	}
	//req.Method = "POST"
	//http.Redirect(w, req, "/user", http.StatusSeeOther)
	err := tpl.ExecuteTemplate(w, "user.gohtml", userdata) //incident{h, u, stdOut})
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}

func user(w http.ResponseWriter, req *http.Request) {
	tf := false
	posh := new()
	user := req.FormValue("username")

	if user == "" {
		em := make(map[string]interface{})
		em["username"] = ""
		em["hostname"] = ""
		err := tpl.ExecuteTemplate(w, "user.gohtml", em)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Fatalln(err)
		}
		return
	}

	ud := pget(posh, "Get-ADUser "+user+" -Properties DisplayName, Department, SamAccountName, Enabled, LockedOut, PasswordExpired, PasswordLastSet, LastBadPasswordAttempt, LastLogonDate, AccountExpirationDate, extensionAttribute13, EmailAddress, Homedirectory | ConvertTo-Json")
	userdata := p2j(ud)
	userdata["username"] = user
	userdata["vpnuser"] = pget(posh, "$t = Get-ADPrincipalGroupMembership "+user+" | Select-Object -Property Name | Where-Object { $_.Name -eq '.LWE-VPN-LWE_User' -or $_.Name -eq '.LWE-VPN-LWE_Admin' }; $t -ne $null")
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

	uf := req.FormValue("UserForm")
	switch uf {
	case "ResetPW":
		fmt.Printf("ResetPW")
		userdata["response"] = pget(posh, "Set-ADAccountPassword -Reset -NewPassword (ConvertTo-SecureString -String 'Liebherr1!' -AsPlainText -Force) -Identity "+user)
		userdata["command"] = "Reset Password"
		tf = true
	case "UnlockAcc":
		fmt.Printf("UnlockAcc")
		userdata["response"] = pget(posh, "Unlock-ADAccount -Identity "+user+"; $?")
		userdata["command"] = "Unlock Account"
		tf = true
	case "ShowPerm":
		fmt.Printf("ShowPerm")
		groups := pget(posh, "Get-ADPrincipalGroupMembership "+user+" | Select-Object -Property Name | Sort-Object -Property Name #| ConvertTo-Json")
		groups = strings.Replace(groups, "\n", "<br>", -1)
		fmt.Printf(groups)
		userdata["response"] = groups
		userdata["command"] = "Show Permissions"
	case "ExtExp":
		fmt.Printf("ExtExp")
		userdata["response"] = pget(posh, "Set-ADUser "+user+" -AccountExpirationDate (Get-Date).AddDays(14); $?")
		userdata["command"] = "Extent Account Expirationdate"
		tf = true
	case "LastHost":
		fmt.Printf("LastHost")
		userdata["response"] = pget(posh, "Unlock-ADAccount -Identity "+user+"; $?")
		userdata["command"] = "Get Last used Computer"
	case "EnableJabber":
		fmt.Printf("Enable Jabber")
		userdata["response"] = exec.Command("\\\\lwesv0170\\itsm_tools\\jabber\\jabber.exe", user)
		userdata["command"] = "Enable Cisco Jabber"
		tf = true
	case "AddVPN":
		fmt.Printf("AddVPN")
		userdata["response"] = pget(posh, "$u = Get-ADUser "+user+" -Properties * ; $g = Get-ADGroup '.LWE-VPN-LWE_User' ; Set-ADObject -Identity $g -Add @{member=$u.DistinguishedName}; $?")
		userdata["command"] = "Add User to VPN-Group"
		tf = true
	case "EmergencyVPN":
		fmt.Printf("EmergencyVPN")
		userdata["response"] = pget(posh, "Unlock-ADAccount -Identity "+user)
		userdata["command"] = "Assign Emergency-VPN"
	}
	if tf == true {
		userdata["response"] = userdata["response"].(string)[:4]
		if userdata["response"].(string) == "True" {
			userdata["response"] = "Success"
		} else {
			userdata["response"] = "Failed"
		}
	}

	err := tpl.ExecuteTemplate(w, "user.gohtml", userdata) //incident{h, u, stdOut})
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}

