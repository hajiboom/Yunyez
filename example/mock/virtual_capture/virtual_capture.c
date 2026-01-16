#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <time.h>

#define RTSP_PORT 8554
#define BUFFER_SIZE 1024
#define MAX_URL_LENGTH 256

// RTSP Methods
typedef enum {
    OPTIONS,
    DESCRIBE,
    SETUP,
    PLAY,
    PAUSE,
    TEARDOWN
} RtspMethod;

// Function to create RTSP request
char* create_rtsp_request(RtspMethod method, const char* url, int cseq, const char* session_id, const char* extra_header) {
    static char buffer[BUFFER_SIZE];
    const char* method_str;
    
    switch(method) {
        case OPTIONS: method_str = "OPTIONS"; break;
        case DESCRIBE: method_str = "DESCRIBE"; break;
        case SETUP: method_str = "SETUP"; break;
        case PLAY: method_str = "PLAY"; break;
        case PAUSE: method_str = "PAUSE"; break;
        case TEARDOWN: method_str = "TEARDOWN"; break;
        default: method_str = "OPTIONS"; break;
    }
    
    if (session_id && strlen(session_id) > 0) {
        if (extra_header && strlen(extra_header) > 0) {
            snprintf(buffer, sizeof(buffer),
                "%s %s RTSP/1.0\r\n"
                "CSeq: %d\r\n"
                "Session: %s\r\n"
                "%s\r\n"
                "\r\n",
                method_str, url, cseq, session_id, extra_header);
        } else {
            snprintf(buffer, sizeof(buffer),
                "%s %s RTSP/1.0\r\n"
                "CSeq: %d\r\n"
                "Session: %s\r\n"
                "\r\n",
                method_str, url, cseq, session_id);
        }
    } else {
        if (extra_header && strlen(extra_header) > 0) {
            snprintf(buffer, sizeof(buffer),
                "%s %s RTSP/1.0\r\n"
                "CSeq: %d\r\n"
                "%s\r\n"
                "\r\n",
                method_str, url, cseq, extra_header);
        } else {
            snprintf(buffer, sizeof(buffer),
                "%s %s RTSP/1.0\r\n"
                "CSeq: %d\r\n"
                "\r\n",
                method_str, url, cseq);
        }
    }
    
    return buffer;
}

int main(int argc, char *argv[]) {
    int sock;
    struct sockaddr_in server_addr;
    char buffer[BUFFER_SIZE];
    int cseq = 1;
    char session_id[64] = "";
    char *request;
    ssize_t bytes_sent, bytes_received;
    
    // Default values
    char server_ip[16] = "127.0.0.1";
    char stream_url[MAX_URL_LENGTH] = "rtsp://127.0.0.1:8554/mystream";
    
    // Parse command line arguments
    if (argc >= 2) {
        strncpy(server_ip, argv[1], sizeof(server_ip) - 1);
    }
    if (argc >= 3) {
        snprintf(stream_url, sizeof(stream_url), "rtsp://%s:8554/%s", server_ip, argv[2]);
    }
    
    printf("Connecting to RTSP server at %s\n", server_ip);
    printf("Using stream URL: %s\n", stream_url);
    
    // Create socket
    sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        perror("Could not create socket");
        return 1;
    }
    
    // Configure server address
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(RTSP_PORT);
    server_addr.sin_addr.s_addr = inet_addr(server_ip);
    
    // Connect to server
    if (connect(sock, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        perror("Could not connect to server");
        close(sock);
        return 1;
    }
    
    printf("Connected to RTSP server\n");
    
    // Step 1: Send OPTIONS request
    printf("\nStep 1: Sending OPTIONS request...\n");
    request = create_rtsp_request(OPTIONS, stream_url, cseq++, NULL, NULL);
    printf("Sending: %s", request);
    
    bytes_sent = send(sock, request, strlen(request), 0);
    if (bytes_sent < 0) {
        perror("Send failed");
        close(sock);
        return 1;
    }
    
    // Receive response
    memset(buffer, 0, sizeof(buffer));
    bytes_received = recv(sock, buffer, sizeof(buffer) - 1, 0);
    if (bytes_received < 0) {
        perror("Receive failed");
        close(sock);
        return 1;
    }
    buffer[bytes_received] = '\0';
    printf("Received: %s\n", buffer);
    
    // Step 2: Send DESCRIBE request
    printf("Step 2: Sending DESCRIBE request...\n");
    request = create_rtsp_request(DESCRIBE, stream_url, cseq++, NULL, "Accept: application/sdp");
    printf("Sending: %s", request);
    
    bytes_sent = send(sock, request, strlen(request), 0);
    if (bytes_sent < 0) {
        perror("Send failed");
        close(sock);
        return 1;
    }
    
    // Receive response
    memset(buffer, 0, sizeof(buffer));
    bytes_received = recv(sock, buffer, sizeof(buffer) - 1, 0);
    if (bytes_received < 0) {
        perror("Receive failed");
        close(sock);
        return 1;
    }
    buffer[bytes_received] = '\0';
    printf("Received: %s\n", buffer);
    
    // Step 3: Send SETUP request
    printf("Step 3: Sending SETUP request...\n");
    request = create_rtsp_request(SETUP, strcat(strcpy(buffer, stream_url), "/track1"), 
                                 cseq++, NULL, "Transport: RTP/AVP;unicast;client_port=8000-8001");
    printf("Sending: %s", request);
    
    bytes_sent = send(sock, request, strlen(request), 0);
    if (bytes_sent < 0) {
        perror("Send failed");
        close(sock);
        return 1;
    }
    
    // Receive response
    memset(buffer, 0, sizeof(buffer));
    bytes_received = recv(sock, buffer, sizeof(buffer) - 1, 0);
    if (bytes_received < 0) {
        perror("Receive failed");
        close(sock);
        return 1;
    }
    buffer[bytes_received] = '\0';
    printf("Received: %s\n", buffer);
    
    // Extract session ID from response
    char* session_header = strstr(buffer, "Session:");
    if (session_header) {
        session_header += strlen("Session:");
        // Skip any leading spaces
        while (*session_header == ' ') session_header++;
        
        char* end = strchr(session_header, ';');
        if (!end) end = strchr(session_header, '\r');
        if (!end) end = strchr(session_header, '\n');
        
        if (end) {
            int len = end - session_header;
            strncpy(session_id, session_header, len);
            session_id[len] = '\0';
            
            // If there's a semicolon, we might have additional params like timeout
            char* semicolon = strchr(session_id, ';');
            if (semicolon) {
                *semicolon = '\0';
            }
            
            printf("Extracted Session ID: %s\n", session_id);
        }
    }
    
    // Step 4: Send PLAY request
    printf("Step 4: Sending PLAY request...\n");
    request = create_rtsp_request(PLAY, stream_url, cseq++, session_id, NULL);
    printf("Sending: %s", request);
    
    bytes_sent = send(sock, request, strlen(request), 0);
    if (bytes_sent < 0) {
        perror("Send failed");
        close(sock);
        return 1;
    }
    
    // Receive response
    memset(buffer, 0, sizeof(buffer));
    bytes_received = recv(sock, buffer, sizeof(buffer) - 1, 0);
    if (bytes_received < 0) {
        perror("Receive failed");
        close(sock);
        return 1;
    }
    buffer[bytes_received] = '\0';
    printf("Received: %s\n", buffer);
    
    // Simulate playing for a few seconds
    printf("Playing stream for 5 seconds...\n");
    sleep(5);
    
    // Step 5: Send TEARDOWN request
    printf("Step 5: Sending TEARDOWN request...\n");
    request = create_rtsp_request(TEARDOWN, stream_url, cseq++, session_id, NULL);
    printf("Sending: %s", request);
    
    bytes_sent = send(sock, request, strlen(request), 0);
    if (bytes_sent < 0) {
        perror("Send failed");
        close(sock);
        return 1;
    }
    
    // Receive response
    memset(buffer, 0, sizeof(buffer));
    bytes_received = recv(sock, buffer, sizeof(buffer) - 1, 0);
    if (bytes_received < 0) {
        perror("Receive failed");
        close(sock);
        return 1;
    }
    buffer[bytes_received] = '\0';
    printf("Received: %s\n", buffer);
    
    // Close socket
    close(sock);
    printf("Connection closed.\n");
    
    return 0;
}