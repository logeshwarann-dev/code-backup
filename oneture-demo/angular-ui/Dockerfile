# Stage 1: Build the Angular application
FROM node:20-alpine AS build

# Setup the working directory
WORKDIR /usr/src/app

# Copy package.json and package-lock.json to install dependencies first (for better caching)
COPY package.json package-lock.json ./

# Install dependencies (without copying node_modules)
RUN npm ci

RUN npm install -g @angular/cli

# Copy the entire project (excluding node_modules due to .dockerignore)
COPY . .

# Build Angular application in PROD mode
RUN npm run build --configuration=production

# Stage 2: Serve the application using Nginx
FROM nginx:alpine

# Install OpenSSL to generate self-signed certs
# RUN apk add --no-cache openssl

# Generate a self-signed SSL certificate
# RUN mkdir -p /etc/nginx/ssl && \
#     openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
#     -keyout /etc/nginx/ssl/nginx.key \
#     -out /etc/nginx/ssl/nginx.crt \
#     -subj "/C=US/ST=SomeState/L=SomeCity/O=MyOrg/CN=localhost"


    # Copy Nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built Angular files
COPY --from=build /usr/src/app/dist/bse-frontend/browser /usr/share/nginx/html

EXPOSE 443

CMD ["nginx", "-g", "daemon off;"]