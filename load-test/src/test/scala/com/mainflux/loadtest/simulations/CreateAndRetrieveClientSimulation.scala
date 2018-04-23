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
import CreateAndRetrieveClientSimulation._
import io.gatling.http.protocol.HttpProtocolBuilder.toHttpProtocol
import io.gatling.http.request.builder.HttpRequestBuilder.toActionBuilder
import com.mainflux.loadtest.simulations.UrlConstants.ManagerUrl

class CreateAndRetrieveClientSimulation extends Simulation {

  // Register user
  Http(s"${ManagerUrl}/users")
    .postData(User)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString

  // Login user
  val tokenRes = Http(s"${ManagerUrl}/tokens")
    .postData(User)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
    .body

  val tokenCursor = parse(tokenRes).getOrElse(Json.Null).hcursor
  val token = tokenCursor.downField("token").as[String].getOrElse("")

  // Prepare testing scenario
  val httpProtocol = http
    .baseURL(ManagerUrl)
    .inferHtmlResources()
    .acceptHeader("*/*")
    .contentTypeHeader(ContentType)
    .userAgentHeader("curl/7.54.0")

  val scn = scenario("CreateAndGetClient")
    .exec(http("request_0")
      .post("/clients")
      .header(HttpHeaderNames.ContentType, ContentType)
      .header(HttpHeaderNames.Authorization, token)
      .body(StringBody(Client))
      .check(status.is(201))
      .check(headerRegex(HttpHeaderNames.Location, "(.*)").saveAs("location")))
    .exec(http("request_1")
      .get("${location}")
      .header(HttpHeaderNames.Authorization, token)
      .check(status.is(200)))

  setUp(
    scn.inject(
      constantUsersPerSec(100) during (15 second),
      nothingFor(15 second),
      constantUsersPerSec(250) during (15 second),
      nothingFor(15 second),
      constantUsersPerSec(500) during (15 second),
      nothingFor(15 second),
      constantUsersPerSec(750) during (15 second),
      nothingFor(15 second),
      constantUsersPerSec(1000) during (15 second))).protocols(httpProtocol)
}

object CreateAndRetrieveClientSimulation {
  val ContentType = "application/json; charset=utf-8"
  val User = """{"email":"john.doe@email.com", "password":"123"}"""
  val Client = """{"type":"device", "name":"weio"}"""
}