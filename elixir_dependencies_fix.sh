brew uninstall elixir
brew uninstall erlang
brew install erlang@25
brew install --ignore-dependencies elixir
echo 'export PATH="/usr/local/opt/erlang@25/bin:$PATH"' >> ~/.zshrc



curl -fsSO https://elixir-lang.org/install.sh
sh install.sh elixir@1.18.2 otp@25.1.2
installs_dir=$HOME/.elixir-install/installs
export PATH=$installs_dir/otp/25.1.2/bin:$PATH
export PATH=$installs_dir/elixir/1.18.2-otp-25/bin:$PATH
iex

export PATH=$HOME/.elixir-install/installs/otp/25.1.2/bin:$PATH
export PATH=$HOME/.elixir-install/installs/elixir/1.18.2-otp-25/bin:$PATH