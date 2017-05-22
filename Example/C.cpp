#include "ifstream"

int main() {

    std::ifstream meta_ifs;
    meta_ifs.open(meta_file.c_str());
    std::stringstream ss;
    ss << meta_ifs.rdbuf();
    meta_ifs.close();
    std::string meta_data_cpp = ss.str();
    char const *meta_data = meta_data_cpp.c_str();
    int data_length = strlen(meta_data);
    unsigned int encoded_data_length = Base64encode_len(data_length); //include a null terminator
    char *base64_meta_data = (char *) malloc(encoded_data_length);
    Base64encode(base64_meta_data, meta_data, data_length);
    WebSocket *m_psock = new WebSocket(cs, request, response);
    m_psock->setReceiveTimeout(Poco::Timespan(600, 0));
    int bigendian = htonl(encoded_data_length - 1);
    m_psock->sendFrame(&bigendian, 4, WebSocket::FRAME_BINARY);
    int len = m_psock->sendFrame(base64_meta_data, encoded_data_length - 1, WebSocket::FRAME_BINARY);
//std::cout << "Sent bytes " << len << std::endl;
    free(base64_meta_data);

    if (!infilename.empty()) {
        short buf[CHUNK_SIZE];
        size_t out_size;
        FILE *fp = fopen(infilename.c_str(), "rb");
        while (1) {
            out_size = fread(buf, sizeof(*buf), CHUNK_SIZE, fp);
            if (out_size == 0) break;
            len = m_psock->sendFrame(buf, out_size * 2, WebSocket::FRAME_BINARY);
        }
        fclose(fp);
    }

    char eos[] = {0x45, 0x4f, 0x53};
// 0x45; 'E'
// 0x4f; 'O'
// 0x53; 'S'
    len = m_psock->sendFrame(eos, 3, WebSocket::FRAME_BINARY);

    int flags = 0;
    int max_recv_size = 1024 * 1024;
    char *recv_buf = (char *) malloc(max_recv_size);
    int recv_len = m_psock->receiveFrame(recv_buf, max_recv_size, flags);
    recv_buf[recv_len] = 0;
    std::ofstream ofs;
    ofs.open(outfilename.c_str());
//skip the first 4 bytes because it's header size
    ofs << ExtractResultFromBase64Response(recv_buf + 4);
    ofs.close();
    m_psock->close();
    free(recv_buf);
}