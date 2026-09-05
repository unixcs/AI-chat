FROM node:22-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install --production --no-audit --no-fund \
    --registry=https://registry.npmmirror.com --replace-registry-host=always

COPY server.js ./
COPY db.js ./
COPY auth.js ./
COPY admin-member-expire.js ./
COPY prompts ./prompts

EXPOSE 3001

CMD ["node", "server.js"]
