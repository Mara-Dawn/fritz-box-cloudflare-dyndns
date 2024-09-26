import os
import CloudFlare
import waitress
import flask
import logging
import datetime

app = flask.Flask(__name__)


@app.route("/", methods=["GET"])
def main():
    token = flask.request.args.get("token")
    zone = flask.request.args.get("zone")
    record = flask.request.args.get("record")
    records = flask.request.args.get("records")
    ipv4 = flask.request.args.get("ipv4")
    ipv6 = flask.request.args.get("ipv6")
    cf = CloudFlare.CloudFlare(token=token)

    app.logger.info(f"request recieved on {datetime.datetime.now()}.")

    if not token:
        app.logger.error("Missing token URL parameter.")
        return (
            flask.jsonify(
                {"status": "error", "message": "Missing token URL parameter."}
            ),
            400,
        )
    if not zone:
        app.logger.error("Missing zone URL parameter.")
        return (
            flask.jsonify(
                {"status": "error", "message": "Missing zone URL parameter."}
            ),
            400,
        )
    if not record and not records:
        app.logger.error("Missing record or records URL parameter.")
        return (
            flask.jsonify(
                {
                    "status": "error",
                    "message": "Missing record or records URL parameter.",
                }
            ),
            400,
        )
    if not ipv4 and not ipv6:
        app.logger.error("Missing ipv4 or ipv6 URL parameter.")
        return (
            flask.jsonify(
                {"status": "error", "message": "Missing ipv4 or ipv6 URL parameter."}
            ),
            400,
        )

    zones = cf.zones.get(params={"name": zone})

    if not zones:
        app.logger.error("Zone {} does not exist.".format(zone))
        return (
            flask.jsonify(
                {"status": "error", "message": "Zone {} does not exist.".format(zone)}
            ),
            404,
        )

    # app.logger.info(f"token = {token}")
    app.logger.info(f"zone = {zone}")
    app.logger.info(f"record = {record}")
    app.logger.info(f"records = {records}")
    app.logger.info(f"ipv4 = {ipv4}")
    app.logger.info(f"ipv6 = {ipv6}")

    if record:
        put_dns_record(cf, zones, record, zone, ipv4, ipv6)
    if records:
        records = records.split(",")
        put_dns_record(cf, zones, None, zone, ipv4, ipv6)
        for r in records:
            put_dns_record(cf, zones, r, zone, ipv4, ipv6)

    app.logger.info("Update finised.")
    app.logger.info("############################")
    return flask.jsonify({"status": "success", "message": "Update successful."}), 200

def put_dns_record(cf, zones, record, zone, ipv4, ipv6):
    if record is not None:
        url = "{}.{}".format(record, zone)
    else:
        url = zone
    app.logger.info(f"Checking changes for {url}")

    try:
        a_record = cf.zones.dns_records.get(
            zones[0]["id"],
            params={"name": url, "match": "all", "type": "A"},
        )
        # app.logger.info(f"A record: {a_record}")
        aaaa_record = cf.zones.dns_records.get(
            zones[0]["id"],
            params={
                "name": url,
                "match": "all",
                "type": "AAAA",
            },
        )
        # app.logger.info(f"AAAA record: {aaaa_record}")

        if ipv4 and not a_record:
            app.logger.error("A record for {} does not exist.".format(url))
            return (
                flask.jsonify(
                    {
                        "status": "error",
                        "message": "A record for {} does not exist.".format(url),
                    }
                ),
                404,
            )

        if ipv6 and not aaaa_record:
            app.logger.error("AAAA record for {} does not exist.".format(url))
            return (
                flask.jsonify(
                    {
                        "status": "error",
                        "message": "AAAA record for {} does not exist.".format(url),
                    }
                ),
                404,
            )

        if ipv4 and a_record[0]["content"] != ipv4:
            app.logger.info(f"Updating A record for {url}: {a_record[0]["content"]} -> {ipv4}")
            cf.zones.dns_records.put(
                zones[0]["id"],
                a_record[0]["id"],
                data={
                    "name": a_record[0]["name"],
                    "type": "A",
                    "content": ipv4,
                    "proxied": a_record[0]["proxied"],
                    "ttl": a_record[0]["ttl"],
                },
            )

        if ipv6 and aaaa_record[0]["content"] != ipv6:
            app.logger.info(f"Updating AAAA record for {url}: {aaaa_record[0]["content"]} -> {ipv6}")
            cf.zones.dns_records.put(
                zones[0]["id"],
                aaaa_record[0]["id"],
                data={
                    "name": aaaa_record[0]["name"],
                    "type": "AAAA",
                    "content": ipv6,
                    "proxied": aaaa_record[0]["proxied"],
                    "ttl": aaaa_record[0]["ttl"],
                },
            )
    except CloudFlare.exceptions.CloudFlareAPIError as e:
        app.logger.error(f"CF API Error {str(e)}")
        return flask.jsonify({"status": "error", "message": str(e)}), 500


app.secret_key = os.urandom(24)
app.logger.handlers.clear()
app.logger.setLevel(logging.DEBUG)
waitress.serve(app, host="0.0.0.0", port=8070)
