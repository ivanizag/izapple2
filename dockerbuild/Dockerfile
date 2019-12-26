FROM circleci/golang:1.12

LABEL MAINTAINER="Ivan Izaguirre <ivanizag@gmail.com>"

RUN sudo apt-get update
RUN sudo apt-get install -y libsdl2-dev mingw-w64

RUN wget https://www.libsdl.org/release/SDL2-devel-2.0.9-mingw.tar.gz
RUN sudo tar -xzf SDL2-devel-2.0.9-mingw.tar.gz
RUN sudo cp -r SDL2-2.0.9/x86_64-w64-mingw32 /usr

COPY buildindocker.sh .
RUN sudo chmod +x buildindocker.sh

CMD ["./buildindocker.sh"]
