docker buildx build --platform linux/arm64 -f .scripts/Dockerfile -t 192.168.1.19:5000/demo_service:latest demo_service/ 
docker push 192.168.1.19:5000/demo_service:latest
