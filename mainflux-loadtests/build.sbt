name := "mainflux-loadtests"
version := "1.0-SNAPSHOT"

enablePlugins(GatlingPlugin)

scalaVersion := "2.12.4"

val gatlingVersion = "2.3.1"

libraryDependencies ++= Seq(
  "io.gatling.highcharts" % "gatling-charts-highcharts" % gatlingVersion % "test,it",
  "io.gatling"            % "gatling-test-framework"    % "2.3.1" % "test,it",
  "org.scalaj"            % "scalaj-http_2.12"          % "2.3.0" % "test,it",
  "com.typesafe.play"     % "play-json_2.12" % "2.6.9"  % "test,it"
)