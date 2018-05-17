package com.mainflux.loadtest.simulations

import scala.concurrent.duration._
import scalaj.http.Http
import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.jdbc.Predef._
import io.circe._
import io.circe.generic.auto._
import io.circe.parser._
import io.circe.syntax._
import PublishMessages._
import io.gatling.http.protocol.HttpProtocolBuilder.toHttpProtocol
import io.gatling.http.request.builder.HttpRequestBuilder.toActionBuilder
import com.mainflux.loadtest.simulations.Constants._

final class PublishMessages extends Simulation {
  Http(s"$UsersURL/users")
    .postData(User)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString

  private val token = {
    val res = Http(s"$UsersURL/tokens")
      .postData(User)
      .header(HttpHeaderNames.ContentType, ContentType)
      .asString
      .body

    val cursor = parse(res).getOrElse(Json.Null).hcursor
    cursor.downField("token").as[String].getOrElse("")
  }

  private val thingID = Http(s"$ThingsURL/clients")
    .postData(Client)
    .header(HttpHeaderNames.Authorization, token)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
    .headers("Location")(0).split("/")(2)

  private val thingAccessKey = {
    val res = Http(s"$ThingsURL/clients/$thingID")
      .header(HttpHeaderNames.Authorization, token)
      .header(HttpHeaderNames.ContentType, ContentType)
      .asString
      .body

    val cursor = parse(res).getOrElse(Json.Null).hcursor
    cursor.downField("key").as[String].getOrElse("")
  }

  private val chanID = Http(s"$ThingsURL/channels")
    .postData(Channel)
    .header(HttpHeaderNames.Authorization, token)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
    .headers("Location")(0)
    .split("/")(2)

  Http(s"$ThingsURL/channels/$chanID/things/$thingID")
    .method("PUT")
    .header(HttpHeaderNames.Authorization, token)
    .asString

  private val httpProtocol = http
    .baseURL(HttpAdapterURL)
    .inferHtmlResources()
    .acceptHeader("*/*")
    .contentTypeHeader("application/json; charset=utf-8")
    .userAgentHeader("curl/7.54.0")

  private val scn = scenario("PublishMessage")
    .exec(http("PublishMessageRequest")
      .post(s"/channels/$chanID/messages")
      .header(HttpHeaderNames.ContentType, "application/senml+json")
      .header(HttpHeaderNames.Authorization, thingAccessKey)
      .body(StringBody(Message))
      .check(status.is(202)))

  setUp(scn.inject(constantUsersPerSec(RequestsPerSecond) during 15.seconds)).protocols(httpProtocol)
}

object PublishMessages {
  val ContentType = "application/json"
  val User = """{"email":"john.doe@email.com", "password":"123"}"""
  val Client = """{"type":"device", "name":"weio"}"""
  val Channel = """{"name":"mychan"}"""
  val Message = """[{"bn":"some-base-name:","bt":1.276020076001e+09, "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}, {"n":"current","t":-5,"v":1.2}, {"n":"current","t":-4,"v":1.3}]"""
}
