#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include "MQTTClient.h"
#include "voice_proto.h"

#define ADDRESS     "tcp://127.0.0.1:1883"
#define CLIENTID    "bridge_mqtt_source_1"
#define USERNAME    "emqxpp"
#define PASSWORD    "emqxpp"
#define TOPIC       "test/T0001/A0001/voice/server"

// Simple CRC16-CCITT (0x1021 polynomial, initial 0xFFFF)
uint16_t crc16_ccitt(const uint8_t *data, size_t len) {
    uint16_t crc = 0xFFFF;
    for (size_t i = 0; i < len; i++) {
        crc ^= (uint16_t)data[i] << 8;
        for (int j = 0; j < 8; j++) {
            if (crc & 0x8000) {
                crc = (crc << 1) ^ 0x1021;
            } else {
                crc <<= 1;
            }
        }
    }
    return crc;
}

int main(int argc, char* argv[]) {
    MQTTClient client;
    MQTTClient_connectOptions conn_opts = MQTTClient_connectOptions_initializer;
    MQTTClient_message pub_msg = MQTTClient_message_initializer;
    MQTTClient_deliveryToken token;
    int rc;

    // Initialize MQTT client
    if ((rc = MQTTClient_create(&client, ADDRESS, CLIENTID,
                                MQTTCLIENT_PERSISTENCE_NONE, NULL)) != MQTTCLIENT_SUCCESS) {
        printf("Failed to create client, return code %d\n", rc);
        exit(EXIT_FAILURE);
    }

    conn_opts.keepAliveInterval = 20;
    conn_opts.cleansession = 1;
    conn_opts.username = USERNAME;
    conn_opts.password = PASSWORD;

    if ((rc = MQTTClient_connect(client, &conn_opts)) != MQTTCLIENT_SUCCESS) {
        printf("Failed to connect, return code %d\n", rc);
        exit(EXIT_FAILURE);
    }

    printf("Connected to MQTT broker at %s\n", ADDRESS);

    // Simulate audio payload
    const char* payload_str = "OPUS_DUMMY_FRAME_1234567890";
    size_t payload_len = strlen(payload_str);
    const uint8_t* payload = (const uint8_t*)payload_str;

    // Build header
    VoiceHeader hdr = {0};
    hdr.version = 1;
    hdr.audio_type = 3;        // OPUS
    hdr.sample_rate = 16000;
    hdr.ch = 1;                // mono
    hdr.f = 1;                 // full frame
    hdr.frame_seq = 1;
    hdr.timestamp = 1000;
    hdr.payload_len = (uint16_t)payload_len;

    // Step 1: pack header (without CRC)
    uint8_t header_buf[HEADER_SIZE];
    voice_header_pack(&hdr, header_buf);
    // Zero out CRC field for checksum
    header_buf[10] = 0;
    header_buf[11] = 0;

    // Step 2: compute CRC over header + payload
    size_t total_len = HEADER_SIZE + payload_len;
    uint8_t* checksum_data = malloc(total_len);
    memcpy(checksum_data, header_buf, HEADER_SIZE);
    memcpy(checksum_data + HEADER_SIZE, payload, payload_len);
    uint16_t computed_crc = crc16_ccitt(checksum_data, total_len);
    free(checksum_data);

    // Step 3: set CRC and repack
    hdr.crc16 = computed_crc;
    voice_header_pack(&hdr, header_buf);

    // Step 4: build full packet
    uint8_t* full_packet = malloc(total_len);
    memcpy(full_packet, header_buf, HEADER_SIZE);
    memcpy(full_packet + HEADER_SIZE, payload, payload_len);

    // Publish
    pub_msg.payload = full_packet;
    pub_msg.payloadlen = (int)total_len;
    pub_msg.qos = 1;
    pub_msg.retained = 0;

    if ((rc = MQTTClient_publishMessage(client, TOPIC, &pub_msg, &token)) != MQTTCLIENT_SUCCESS) {
        printf("Failed to publish message, return code %d\n", rc);
    } else {
        printf("Published to topic '%s' (%d bytes)\n", TOPIC, (int)total_len);
        MQTTClient_waitForCompletion(client, token, 1000);
    }

    free(full_packet);
    MQTTClient_disconnect(client, 10000);
    MQTTClient_destroy(&client);
    return rc == MQTTCLIENT_SUCCESS ? EXIT_SUCCESS : EXIT_FAILURE;
}