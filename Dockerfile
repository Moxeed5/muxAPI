FROM golang:latest

#folder inside of image
WORKDIR /APP

#copy the source code we wrote
COPY . .

#download our dependecies 
RUN go get -d -v ./...

RUN go build -o api .

#EXPOSE THE PORT
EXPOSE 8080

#runs our executable 
CMD ["./api"]