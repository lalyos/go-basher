#PATH=/usr/local/Cellar/bash/4.3.30/bin/:/usr/gnu/bin:/usr/local/bin:/bin:/usr/bin:.

main() {
  # PATH=/usr/local/Cellar/bash/4.3.30/bin/:/usr/gnu/bin:/usr/local/bin:/bin:/usr/bin:.
  # echo PATH=$PATH
  # echo ===
  # /usr/local/bin/which env
  # echo ===
  # echo which env=$(/usr/local/bin/which env)
  # echo which bash=$(/usr/local/bin/which bash)
  set
  ps -ef >bash_stack_trace.$$
  echo BASH_VERSION=$BASH_VERSION
}
