docker build . -t url-analyzer-image

docker run --name url-analyzer-image-container -d -p 8080:8080 url-analyzer-image:latest