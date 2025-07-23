FROM debian:12.11-slim

# 環境変数の設定
ENV DEBIAN_FRONTEND=noninteractive \
    NODE_VERSION=22 \
    GO_VERSION=1.24.5 \
    PATH="/usr/local/go/bin:/root/.local/bin:$PATH" \
    GOPATH="/go" \
    GOROOT="/usr/local/go"

# 基本パッケージとツールのインストール
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    vim \
    nano \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Node.js (npm含む) のインストール
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# npmの最新版にアップデート
RUN npm install -g npm@latest

# @google/gemini-cli のグローバルインストール
RUN npm install -g @google/gemini-cli

# Go のインストール
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-arm64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-arm64.tar.gz \
    && rm go${GO_VERSION}.linux-arm64.tar.gz

# Docker CLIのインストール
RUN install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && chmod a+r /etc/apt/keyrings/docker.gpg \
    && echo \
    "deb [arch=\"$(dpkg --print-architecture)\" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
    \"$(. /etc/os-release && echo \"$VERSION_CODENAME\")\" stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y docker-ce-cli

# Go のワークスペース作成
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin ${GOPATH}/pkg/mod
RUN chmod -R 755 ${GOPATH}

# 開発用ユーザーの作成（rootでの作業を避ける）
RUN useradd -m -s /bin/bash developer \
    && usermod -aG sudo developer \
    && mkdir -p /home/developer/workspace \
    && chown -R developer:developer /home/developer \
    && chown -R developer:developer ${GOPATH}

# 開発用ユーザーにGo環境を設定
USER developer
RUN echo 'export PATH="/usr/local/go/bin:$PATH"' >> /home/developer/.bashrc \
    && echo 'export GOPATH="/go"' >> /home/developer/.bashrc \
    && echo 'export GOROOT="/usr/local/go"' >> /home/developer/.bashrc

# 作業ディレクトリの設定
WORKDIR /home/developer/workspace

# バージョン確認用のヘルスチェック
USER root
RUN node --version && npm --version && go version && git --version

# 最終的に開発用ユーザーに切り替え
USER developer

# コンテナ起動時のデフォルトコマンド
CMD ["/bin/bash"]
