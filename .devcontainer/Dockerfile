FROM ubuntu:24.04

RUN apt-get update && DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install git wget vim bash-completion python3 curl

RUN curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | \
    bash -s -- -v 0.13.5

RUN curl https://raw.githubusercontent.com/git/git/master/contrib/completion/git-completion.bash \
    -o ~/.git-completion.bash

RUN <<EOF cat >> ~/.bashrc
git_branch() {
  git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/(\1)/'
}

export PS1="[\u@\h \W]\[\033[00;32m\]\$(git_branch)\[\033[00m\]\$ "

source ~/.git-completion.bash
EOF
