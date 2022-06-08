FROM alpine:3.16.0
WORKDIR /app
COPY ./autograph-backend-controller.exe ./autograph-backend-controller.exe
CMD ./autograph-backend-controller.exe