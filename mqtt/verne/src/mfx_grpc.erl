-module(mfx_grpc).
-behaviour(gen_server).
-export([
    start_link/0,
    init/1,
    send/2,
    handle_call/3,
    handle_cast/2,
    handle_info/2,
    terminate/2
]).

init(_Args) ->
    error_logger:info_msg("mfx_grpc genserver has started (~w)~n", [self()]),

    [{_, GrpcUrl}] = ets:lookup(mfx_cfg, grpc_url),
    {ok, {_, _, GrpcHost, GrpcPort, _, _}} = http_uri:parse(GrpcUrl),
    error_logger:info_msg("mfx_redis host: ~p,  port: ~p", [GrpcHost, GrpcPort]),
    {ok, GrpcConn} = grpc_client:connect(tcp, GrpcHost, GrpcPort),

    ets:insert(mfx_cfg, {grpc_conn, GrpcConn}),
    {ok, []}.

start_link() ->
    gen_server:start_link({local, ?MODULE}, ?MODULE, [], []).

send(Method, Message) ->
    gen_server:call(?MODULE, {send, Method, Message}).


handle_call({send, identify, Message}, _From, _State) ->
    [{grpc_conn, Conn}] = ets:lookup(mfx_cfg, grpc_conn),
    error_logger:info_msg("mfx_grpc message: ~p", [Message]),
    {Status, Result} = internal_client:'IdentifyThing'(Conn, Message, []),
    case Status of
      ok ->
          #{
              grpc_status := 0,
              headers := #{<<":status">> := <<"200">>},
              http_status := HttpStatus,
              result :=
                  #{value := ThingId},
              status_message := <<>>,
              trailers := #{<<"grpc-status">> := <<"0">>}
          } = Result,

          case HttpStatus of
              200 ->
                  {reply, {ok, ThingId}, ok};
              _ ->
                  {reply, {error, HttpStatus}, error}
          end;
      _ ->
          {reply, {error, Status}, error}
  end;
handle_call({send, can_access_by_id, Message}, _From, _State) ->
  [{grpc_conn, Conn}] = ets:lookup(mfx_cfg, grpc_conn),
  error_logger:info_msg("mfx_grpc message: ~p", [Message]),
  {Status, Result} = internal_client:'CanAccessByID'(Conn, Message, []),
  case Status of
    ok ->
      #{
        grpc_status := 0,
        headers := #{
          <<":status">> := <<"200">>,
          <<"content-type">> := <<"application/grpc+proto">>
        },
        http_status := HttpStatus,
        result := #{},
        status_message := <<>>,
        trailers := #{
          <<"grpc-message">> := <<>>,
          <<"grpc-status">> := <<"0">>}
      } = Result,

      case HttpStatus of
          200 ->
              {reply, ok, ok};
          _ ->
              {reply, {error, HttpStatus}, error}
      end;
      
    _ ->
        {reply, {error, Status}, error}
  end.

handle_cast(_Request, State) ->
    {noreply, State}.

handle_info(_Info, State) ->
    {noreply, State}.

terminate(_Reason, _State) ->
    [].

