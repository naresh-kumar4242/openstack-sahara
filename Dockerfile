FROM golang:1.8
#ToDO - Use golang alpine image & also when get time , use multi-build feature of Docker 
MAINTAINER  Naresh <naresh.kumar4242@xenonstack.com>

RUN mkdir /app
WORKDIR /app
COPY . .

# Installing dependancies
RUN go get github.com/parnurzeal/gorequest
RUN go get github.com/gin-gonic/gin
RUN go get github.com/gin-contrib/cors

#Env vars (Different for every OpenStack setup)

#ENV OS_AUTH_URL http://controller:5000/v3/  (We will pass it during container creation process)

ENV OS_PROJECT_ID ce84c3e40d9d43b5b10b7cd6fb5f3c66
ENV OS_PROJECT_NAME admin
ENV OS_USERNAME admin
ENV OS_PASSWORD uCsl-39G3XWDNDSfPlgB

RUN go build -o main .
ENTRYPOINT ["/app/main"] #Using entrypoint instead of cmd in case i need to pass arguments at run-time to main binary 
EXPOSE 8090
