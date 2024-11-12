import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class ServerAtOnceLoadSimulation extends Simulation {

  val httpProtocolEcho = http.baseUrl("http://echo-hello:8080")
  val httpProtocolGin = http.baseUrl("http://gin-hello:8080")

  val scnEcho = scenario("Echo Server Load Test")
    .exec(http("Echo Metrics at once").get("/"))
  val scnGin = scenario("GinHTTP Server Load Test")
    .exec(http("GinHTTP Metrics at once").get("/"))

  setUp(
    scnEcho.inject(
        atOnceUsers(50000),
        nothingFor(5.seconds)
    ).protocols(httpProtocolEcho),
    scnGin.inject(
        atOnceUsers(50000),
        nothingFor(5.seconds)
    ).protocols(httpProtocolGin)
  ).maxDuration(60.seconds)
}