/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

#include <asm/ioctl.h>
#include <linux/ipmi.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>
#include <sys/ioctl.h>
#include <sys/time.h>

#include "linux_inband.h"

#define TIMEOUT 5

typedef enum error_codes {
    ERR_INVALID_BUFF_SIZE = -2,
    ERR_INVALID_CALL = -1,
    IPMI_OK = 0,
    ERR_IPMI_DEVICE_NOT_OPENED = 100,
    ERR_IMPI_COMMAND_NOT_SENT = 200,
    ERR_IPMI_DEVICE_NOT_READY = 300,
    ERR_IPMI_DEVICE_TIMEOUT = 310,
    ERR_IPMI_MESSAGE_NOT_RECEIVED = 320
} ipmi_error_codes_t;

void IPMI_Syserr(struct IpmiStatusInfo *out) {
	out->system_error = errno;
	strerror_r(out->system_error, out->error_str, sizeof(out->error_str));
}


// error codes:
// 0 - ok
// <0 - invalid call
// >0 - errors from OS
int IPMI_BatchCommands(char *device, struct IpmiCommandInput *inputs,
	struct IpmiCommandOutput *outputs, int n, int n_sim, struct IpmiStatusInfo *info) {
	    ipmi_error_codes_t status = IPMI_OK;
		int fd, i, sent = 0, recvd = 0, readyFds;
		struct ipmi_ipmb_addr sendAddr={0}, recvAddr={0};
		struct ipmi_req request={0};
		struct ipmi_recv recv={0};
		struct timeval timeoutSend, timeoutRecv;
		fd_set fdset;
		unsigned char outData[1024] = {0xFF};

		timeoutSend.tv_sec = TIMEOUT;
		timeoutSend.tv_usec = 0;

		timeoutRecv.tv_sec = TIMEOUT;
		timeoutRecv.tv_usec = 0;


		if (!info) {
			status = ERR_INVALID_CALL;
			return status;
		}

		for(i = 0; i < n; i++)
		{
			if (inputs[i].data_len < 2) {
				strcpy(info->error_str, "Supplied buffer too short in msg %d");
				status = ERR_INVALID_BUFF_SIZE;
				return status;
			}
		}

		fd = open(device, O_RDWR);
		if (fd < 0)
		{
			IPMI_Syserr(info);
			status = ERR_IPMI_DEVICE_NOT_OPENED;
			return status;
		}

		int to_receive = n;
		while(to_receive > 0) {
			if(sent < n && (sent-recvd) < n_sim) {

				sendAddr.addr_type = IPMI_IPMB_ADDR_TYPE;
				sendAddr.channel = inputs[sent].channel;
				sendAddr.slave_addr = inputs[sent].slave;
				sendAddr.lun = 0;

				request.addr = (char*)&sendAddr;
				request.addr_len = sizeof(sendAddr);

				request.msgid = sent;
				outputs[request.msgid].is_valid = 0;

				request.msg.netfn = inputs[sent].data[0];
				request.msg.cmd = inputs[sent].data[1];
				request.msg.data = &inputs[sent].data[2];
				request.msg.data_len = inputs[sent].data_len - 2;

				if (ioctl(fd, IPMICTL_SEND_COMMAND, &request) < 0) {
					IPMI_Syserr(info);
					status = ERR_IMPI_COMMAND_NOT_SENT;
					to_receive--;
				}
				sent++;
				continue;
			}
			//if we are at this point some messages are sent

			FD_ZERO(&fdset);
			FD_SET(fd, &fdset);

			if ( (readyFds = select(fd+1, &fdset, NULL, NULL, &timeoutRecv)) < 0) {
				IPMI_Syserr(info);
				close(fd);
				status = ERR_IPMI_DEVICE_NOT_READY;
				return status;
			}

			if (readyFds < 1) {
				strcpy(info->error_str,"Timeout on read select.");
				status = ERR_IPMI_DEVICE_TIMEOUT;
				close(fd);
				return status;
			}

			recv.addr = (char*)&recvAddr;
			recv.addr_len = sizeof(recvAddr);

			recv.msg.data = outData;
			recv.msg.data_len = sizeof(outData);

			if (ioctl(fd, IPMICTL_RECEIVE_MSG_TRUNC, &recv) < 0) {
				IPMI_Syserr(info);
				status = ERR_IPMI_MESSAGE_NOT_RECEIVED;
			}
			else {
			    // using memcpy here results in glibc dependency, so this simple for loop
			    // avoids that
			    for(i = 0; i < recv.msg.data_len; i++) {
			        outputs[recv.msgid].data[i] = recv.msg.data[i];
			    }
			    outputs[recv.msgid].data_len = recv.msg.data_len;
			    outputs[recv.msgid].is_valid = 1;
			}
			to_receive--;
		}
		close(fd);
		return status;
	}

int IPMI_System_BatchCommands(char *device, struct IpmiCommandInput *inputs,
	struct IpmiCommandOutput *outputs, int n, int n_sim, struct IpmiStatusInfo *info) {
	    ipmi_error_codes_t status = IPMI_OK;
		int fd, i, sent = 0, recvd = 0, readyFds;
		struct ipmi_system_interface_addr sendAddr={0}, recvAddr={0};
		struct ipmi_req request={0};
		struct ipmi_recv recv={0};
		struct timeval timeoutSend, timeoutRecv;
		fd_set fdset;
		unsigned char outData[1024] = {0xFF};

		timeoutSend.tv_sec = TIMEOUT;
		timeoutSend.tv_usec = 0;

		timeoutRecv.tv_sec = TIMEOUT;
		timeoutRecv.tv_usec = 0;


		if (!info) {
			status = ERR_INVALID_CALL;
			return status;
		}

		for(i = 0; i < n; i++)
		{
			if (inputs[i].data_len < 2) {
				strcpy(info->error_str, "Supplied buffer too short in msg %d");
				status = ERR_INVALID_BUFF_SIZE;
				return status;
			}
		}

		fd = open(device, O_RDWR);
		if (fd < 0)
		{
			IPMI_Syserr(info);
			status = ERR_IPMI_DEVICE_NOT_OPENED;
			return status;
		}

		int to_receive = n;
		while(to_receive > 0) {
			if(sent < n && (sent-recvd) < n_sim) {

				sendAddr.addr_type = IPMI_SYSTEM_INTERFACE_ADDR_TYPE;
				sendAddr.channel = 0xF;
				sendAddr.lun = 0;

				request.addr = (char*)&sendAddr;
				request.addr_len = sizeof(sendAddr);

				request.msgid = sent;
				outputs[request.msgid].is_valid = 0;

				request.msg.netfn = inputs[sent].data[0];
				request.msg.cmd = inputs[sent].data[1];
				request.msg.data = &inputs[sent].data[2];
				request.msg.data_len = inputs[sent].data_len-2;

				if (ioctl(fd, IPMICTL_SEND_COMMAND, &request) < 0) {
					IPMI_Syserr(info);
					status = ERR_IMPI_COMMAND_NOT_SENT;
					to_receive--;
				}
				sent++;
				continue;
			}
			//if we are at this point some messages are sent

			FD_ZERO(&fdset);
			FD_SET(fd, &fdset);

			if ( (readyFds = select(fd+1, &fdset, NULL, NULL, &timeoutRecv)) < 0) {
				IPMI_Syserr(info);
				close(fd);
				status = ERR_IPMI_DEVICE_NOT_READY;
				return status;
			}

			if (readyFds < 1) {
				strcpy(info->error_str,"Timeout on read select.");
				status = ERR_IPMI_DEVICE_TIMEOUT;
				close(fd);
				return status;
			}

			recv.addr = (char*)&recvAddr;
			recv.addr_len = sizeof(recvAddr);

			recv.msg.data = outData;
			recv.msg.data_len = sizeof(outData);

			if (ioctl(fd, IPMICTL_RECEIVE_MSG_TRUNC, &recv) < 0) {
				IPMI_Syserr(info);
				status = ERR_IPMI_MESSAGE_NOT_RECEIVED;
			}
			else {
			    // using memcpy here results in glibc dependency, so this simple for loop
			    // avoids that
			    for(i = 0; i < recv.msg.data_len; i++) {
			        outputs[recv.msgid].data[i] = recv.msg.data[i];
			    }
			    outputs[recv.msgid].data_len = recv.msg.data_len;
			    outputs[recv.msgid].is_valid = 1;
			}
			to_receive--;
		}
		close(fd);
		return status;
	}