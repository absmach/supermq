var clientKey = "";

function access(s) {
    s.on('upload', function (data) {
        while (data == "") {
            return s.AGAIN
        }

        if (clientKey === "") {
            clientKey = parseCert(s.variables.ssl_client_s_dn, "CN");
        }

        var pass = parsePackage(s, data);

        if (!clientKey.length || pass !== clientKey) {
            s.log("Cert CN (" + clientKey + ") does not match ID");
            s.off('upload')
            s.deny();
        }

        s.off('upload');
        s.allow();
    })
}

function parsePackage(s, data) {
    // An explanation of MQTT packet structure can be found here:
    // https://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html#msg-format. 
    var packet_type_flags_byte = data.codePointAt(0);
    // First MQTT packet contain message type and flags. CONN message type
    // is encoded as 0001, and we're not interested in flags, so all values
    // 0001xxxx are valid for us, which is between 16 and 32.
    if (packet_type_flags_byte >= 16 && packet_type_flags_byte < 32) {
        // Extract variable length header. It's 1-4 bytes. As long as continuation byte is
        // 1, there are more bytes in this header.
        var len_size = 1;
        for (var remaining_len = 1; remaining_len < 5; remaining_len++) {
            if (data.codePointAt(remaining_len) > 128) {
                len_size += 1;
                continue;
            }
            break;
        }
        // CONTROL(1) + MSG_LEN(1-4) + PROTO_NAME_LEN(2) + PROTO_NAME(4) + PROTO_VERSION(1)
        var flags_pos = 1 + len_size + 2 + 4 + 1;
        var flags = data.codePointAt(flags_pos);
        // If there are no username and password flags (11xxxxxx), return.
        if (flags < 192) {
            return "";
        }
        // FLAGS(1) + KEEP_ALIVE(2)
        var shift = flags_pos + 1 + 2;

        var client_id_len_msb = data.codePointAt(shift).toString(16);
        var client_id_len_lsb = data.codePointAt(shift + 1).toString(16);
        var client_id_len = calcLen(client_id_len_msb, client_id_len_lsb);

        shift = shift + 2 + client_id_len;

        var username_len_msb = data.codePointAt(shift).toString(16);
        var username_len_lsb = data.codePointAt(shift + 1).toString(16);
        var username_len = calcLen(username_len_msb, username_len_lsb);

        shift = shift + 2 + username_len;

        var password_len_msb = data.codePointAt(shift).toString(16);
        var password_len_lsb = data.codePointAt(shift + 1).toString(16);
        var password_len = calcLen(password_len_msb, password_len_lsb);

        shift += 2;
        var password = data.substring(shift, shift + password_len);

        return password;
    }

    return "";    
}

function setKey(r) {
    if (clientKey === "") {
        clientKey = parseCert(r.variables.ssl_client_s_dn, "CN");
    }

    return clientKey;
}

function calcLen(msb, lsb) {
    if (lsb < 2) {
        lsb = "0" + lsb;
    }

    return parseInt(msb + lsb, 16);
}

function parseCert(cert, key) {
    if (cert.length) {
        var pairs = cert.split(',');
        for (var i = 0; i < pairs.length; i++) {
            var pair = pairs[i].split('=');
            if (pair[0].toUpperCase() == key) {
                return pair[1];
            }
        }
    }

    return "";
}
