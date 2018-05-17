package com.mainflux.loadtest

import com.mainflux.loadtest.Constants._
import io.circe._
import io.circe.parser._
import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.http.protocol.HttpProtocolBuilder.toHttpProtocol
import io.gatling.http.request.builder.HttpRequestBuilder.toActionBuilder
import scalaj.http.Http

import scala.concurrent.duration._

final class CreateAndRetrieveThings extends Simulation {
  import CreateAndRetrieveThings._

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

  private val httpProtocol = http
    .baseURL(ThingsURL)
    .inferHtmlResources()
    .acceptHeader("*/*")
    .contentTypeHeader(ContentType)
    .userAgentHeader("curl/7.54.0")

  private val scn = scenario("create and retrieve things")
    .exec(http("create thing")
      .post("/clients")
      .header(HttpHeaderNames.ContentType, ContentType)
      .header(HttpHeaderNames.Authorization, token)
      .body(StringBody(Thing))
      .check(status.is(201))
      .check(headerRegex(HttpHeaderNames.Location, "(.*)").saveAs("location")))
    .exec(http("retrieve thing")
      .get("${location}")
      .header(HttpHeaderNames.Authorization, token)
      .check(status.is(200)))

  setUp(scn.inject(constantUsersPerSec(RequestsPerSecond) during 15.seconds)).protocols(httpProtocol)
}

object CreateAndRetrieveThings {
  val ContentType = "application/json"
  val User = """{"email":"john.doe@email.com", "password":"123"}"""
  val Thing = """{"type":"device", "name":"weio"}"""
}