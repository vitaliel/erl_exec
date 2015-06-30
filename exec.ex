defmodule Breakout.Exec do

  def start(cmd, args \\ []) do
    parent = self
    pid = spawn fn -> init(parent, cmd, args) end
    Process.monitor pid

    {:ok, pid}
  end

  def put_input(pid, data) do
    send pid, {:input, data}
  end

  def stop(pid) do
    send pid, :stop
  end

  def init(parent, cmd, args) do
    if exe = :os.find_executable(:erlang.binary_to_list(cmd)) do
      cmd = List.to_string(exe)
    end

    port = Port.open(
      {:spawn_executable, cmd},
      [{:args, args}, {:packet, 2}, :binary, :use_stdio, :exit_status, :hide]
    )
    state = %{port: port, parent: parent}
    loop(state)
  end

  def loop(state) do
    receive do
      {:input, data} ->
        Port.command(state.port, <<0, data::binary>>)
        loop(state)
      :stop ->
        Port.command(state.port, <<1>>)
        # send port, {self, :close}
        loop(state)
      {_from, {:data, <<0, input::binary>>}} ->
        send state.parent, {:data, :out, input}
        loop(state)
      {_from, {:data, <<1, input::binary>>}} ->
        send state.parent, {:data, :err, input}
        loop(state)
      {_from, {:exit_status, status}} ->
        send state.parent, {:exit_status, status}
      msg ->
        IO.puts "unexpected msg: #{inspect msg}"
        loop(state)
    end
  end
end

defmodule Client do
  def run do
    # {:ok, pid} = Breakout.Exec.start("go", ["run", "src/erl_port/cmd/erl_port/main.go", "--",
    #   "echo", "hello"])

    # {:ok, pid} = Breakout.Exec.start("go", ["run", "src/erl_port/cmd/erl_port/main.go", "--",
    #     "git-upload-pack", "/Users/lz/projects/elixir-train/breakout_sshd/.git"])

    {:ok, pid} = Breakout.Exec.start("go", ["run", "src/erl_port/cmd/erl_port/main.go", "--",
      "ruby", "gets.rb"])

    spawn fn ->
      Breakout.Exec.put_input(pid, "Noroc 1\n")
      Breakout.Exec.put_input(pid, "Noroc 2\n")
      Breakout.Exec.stop(pid)
    end

    loop
  end

  def loop do
    receive do
      {:DOWN, _ref, :process, _pid, state} ->
        IO.puts "Port terminated #{state}"
      msg ->
        IO.puts inspect(msg)
        loop
    end
  end
end

Client.run
