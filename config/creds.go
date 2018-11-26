package config

import (
  "encoding/json"
  "fmt"
  "os"
  "strings"
)

const (

// Email server

   ServerName string = "email-smtp.us-west-2.amazonaws.com:465"

// Sender Email & Password (AWS Creds)- we will pass during run-time

// Cluster static Config

   InternalNetworkID string = "4e56fa4d-2868-495c-9166-074699480ec7"
   KeyPair string = "controller-root"

   ClusterTemplateID string = "5fcac67f-7cd3-4c00-9b14-127f7fa2f330"

// Following are the IDs for Hadoop images
   Ubuntu_14_04_ID string = "49da9eee-5bea-48a8-9759-d56cb5f0343e"
   Ubuntu_16_04_ID string = ""
   CentOS_6_9_ID string = ""
   CentOS_7_ID string = "353f5b99-becd-4301-8fd0-2d1b6b4a43ba"

// Following are the IDs for VM flavours & images
   VMFlavSmall string = "2234e0bd-6860-4f0a-87f3-52b334af6a5e"
   VMFlavMedium string = "7f162d27-7ed6-4ad7-b5fe-81d7b2f85be4"
   VMFlavLarge string = "6cd605e0-c611-4c87-9818-fd30b90e7185"

  UbuntuID string = "2f255161-6ad2-4950-871b-d2b5fa747313"
  CirrosID string = "e3c5e3bc-ca00-4479-b618-f3e541beb53c"
  CentOSID string = "9545a142-3af5-47d9-a7f6-5b34307231bf"


  NetworkID string = "5bbefaf4-96b1-4b1c-94fe-a6ce0882f6b6"   //Provider N/W
  ServerPort string = "8090"
)

type AuthInfo struct {
    Auth struct {
        Identity struct {
            Methods []string `json:"methods"`
            Password struct {
                User struct {
                    Domain struct {
                        Name string `json:"name"`
                    } `json:"domain"`
                    Name string `json:"name"`
                    Password string `json:"password"`
                } `json:"user"`
            } `json:"password"`
        } `json:"identity"`
        Scope struct {
            Project struct {
                Domain struct {
                    Name string `json:"name"`
                } `json:"domain"`
                Name string `json:"name"`
            } `json:"project"`
        } `json:"scope"`
    } `json:"auth"`
}

func Token() (string) {
  payload := AuthInfo{}
  payload.Auth.Identity.Methods = []string{"password"}
  payload.Auth.Identity.Password.User.Domain.Name = "Default"
  payload.Auth.Identity.Password.User.Name = os.ExpandEnv("$OS_USERNAME")
  payload.Auth.Identity.Password.User.Password = os.ExpandEnv("$OS_PASSWORD")
  payload.Auth.Scope.Project.Domain.Name = "Default"
  payload.Auth.Scope.Project.Name = os.ExpandEnv("$OS_PROJECT_NAME")
  token_input, err := json.Marshal(&payload)
  if err != nil {
      panic (err)
  }
  return string(token_input)
}

var  identity_api string = os.ExpandEnv("$OS_AUTH_URL")
var s []string  = strings.Split(identity_api, ":")
//Here s provides the slice [ http //controller auth_port ]
var (
  TokenApi string = fmt.Sprintf("%s/auth/tokens",identity_api)
  ComputeApi string = fmt.Sprintf("%s:%s:8774/v2.1/servers",s[0],s[1])
)
