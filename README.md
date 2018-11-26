## Intro :::

- I am using OpenStack VM & sahara REST APIs
- Program communicates with OpenStack Sahara Data Processing API for launching Hadoop cluster.
- It collects all the OpenStack initial configuration from Environment variables.
 - https://12factor.net/config - Following good coding practices is better for Ops folks , right ;)
  - PS - I am a DevOps guy - So i can understand the pain of managing 10-20s of microservices, if not 100    
- Anyways , First we hit the OpenStack identity API for authenticating ourselves & in return we get Authentication token which we will use further for communicating with Sahara API.
 - I am terrible at UI part , so i made only APIs to see how i can automate this process on my OpenStack setup
 - By the way , i used Postman to hit endpoints
- We can take entire cluster config from user(front-end) & finally launch Hadoop cluster according to user's required configuration.
- Golang Gin framework is being used for handling configs from user.
- PS - I'll keep adding new stuff in feature branches (like i just added VM create code) , will push to master after i write test cases (need to learn) & make sure they pass ;)
## USAGE :::

# Case 1 (Deployment on Docker) :
- Please Clone the repo

- Define all the required vars in Dockerfile (UserName,Password,ProjectName,ProjectID,AuthURL)
- or we can pass Env vars during running docker containers (Preferred method)
- Afterwards,Run following commands to build the image from Dockerfile & finally run our server in container :
  ```
  $ docker build -t openstack_sahara:v1.0 .
  $ docker run -itd -p 8090:8090 --name openstack_sahara -e OS_AUTH_URL=http://10.0.0.11:5000/v3/ openstack_sahara:v1.0

  Now we can access API through localhost:8090
  ```
- Now,Our API server is up & running & it will launch hadoop cluster as soon as it will get Config from user.
  ```
# Case 2 (Legacy/VM Deployment) :
- First of all,Go to OpenStack Dashboard > API Access & then Download OpenStack RC File V3.

- Then, Run following command on the server where the main server program will be executed:
    ```
    $ source admin-openrc.sh
    Provide OpenStack admin password for authentication
    ```
- Now that we have added OpenStack config details to our environment , Head on to build
    ```
    $ go build main.go
    $ ./main
    ```

- Now,Our API server is up & running & it will launch hadoop cluster as soon as it will get Config from user.

**- Naresh Kumar (Dr z0x)**
