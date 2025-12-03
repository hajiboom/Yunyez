#ifndef VOICE_PROTOCOL_H
#define VOICE_PROTOCOL_H

#include <stdint.h>

#define HEADER_SIZE 12

typedef struct {
    uint8_t version;
    uint8_t audio_type;
    uint16_t sample_rate;
    uint8_t ch;
    uint8_t f;
    uint16_t frame_seq;
    uint16_t timestamp;
    uint16_t payload_len;
    uint16_t crc16;
} VoiceHeader;

void voice_header_pack(const VoiceHeader* hdr, uint8_t* buf) {
    buf[0] = ((hdr->version & 0x0F) << 4) | ((hdr->audio_type >> 4) & 0x0F);
    buf[1] = ((hdr->audio_type & 0x0F) << 4) | ((hdr->sample_rate >> 12) & 0x0F);
    buf[2] = (hdr->sample_rate >> 4) & 0xFF;
    buf[3] = ((hdr->sample_rate & 0x0F) << 4) | ((hdr->ch & 0x03) << 2) | (hdr->f & 0x03);
    buf[4] = (hdr->frame_seq >> 8) & 0xFF;
    buf[5] = hdr->frame_seq & 0xFF;
    buf[6] = (hdr->timestamp >> 8) & 0xFF;
    buf[7] = hdr->timestamp & 0xFF;
    buf[8] = (hdr->payload_len >> 8) & 0xFF;
    buf[9] = hdr->payload_len & 0xFF;
    buf[10] = (hdr->crc16 >> 8) & 0xFF;
    buf[11] = hdr->crc16 & 0xFF;
}

void voice_header_unpack(const uint8_t* buf, VoiceHeader* hdr) {
    hdr->version = (buf[0] >> 4) & 0x0F;
    hdr->audio_type = ((buf[0] & 0x0F) << 4) | (buf[1] >> 4);
    hdr->sample_rate = ((uint16_t)(buf[1] & 0x0F) << 12) |
                       ((uint16_t)buf[2] << 4) |
                       ((buf[3] >> 4) & 0x0F);
    hdr->ch = (buf[3] >> 2) & 0x03;
    hdr->f = buf[3] & 0x03;
    hdr->frame_seq = ((uint16_t)buf[4] << 8) | buf[5];
    hdr->timestamp = ((uint16_t)buf[6] << 8) | buf[7];
    hdr->payload_len = ((uint16_t)buf[8] << 8) | buf[9];
    hdr->crc16 = ((uint16_t)buf[10] << 8) | buf[11];
}

#endif