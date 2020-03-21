NAME = paralleli

$(NAME): *.go
	go build -o $(NAME) .

install:
	go install
