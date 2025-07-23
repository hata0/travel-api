FROM debian:12.11-slim

# 環境変数の設定
ENV DEBIAN_FRONTEND=noninteractive \
    NODE_VERSION=22 \
    GO_VERSION=1.24.5 \
    PATH="/usr/local/go/bin:/root/.local/bin:$PATH" \
    GOPATH="/go" \
    GOROOT="/usr/local/go"

# 基本パッケージとツールのインストール
RUN apt-get update && apt-get install -y     curl     wget     git     build-essential     ca-certificates     gnupg     lsb-release     vim     nano     unzip     && rm -rf /var/lib/apt/lists/*     && dpkg --add-architecture amd64     && apt-get update     && apt-get install -y libc6-dev:amd64

# Node.js (npm含む) のインストール
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# npmの最新版にアップデート
RUN npm install -g npm@latest

# @google/gemini-cli のグローバルインストール
RUN npm install -g @google/gemini-cli

# Go のインストール
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

# Go のワークスペース作成
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin ${GOPATH}/pkg
RUN chmod -R 755 ${GOPATH}

# 開発用ユーザーの作成（rootでの作業を避ける）
RUN useradd -m -s /bin/bash developer     && usermod -aG sudo developer     && mkdir -p /home/developer/workspace     && chown -R developer:developer /home/developer     && chown -R developer:developer /go

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
