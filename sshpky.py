#!/usr/bin/python
# -*- coding: UTF-8 -*-

""" Automate SSH logins when you're force to authenticate with a password. """
import getpass
import optparse
import os
import sys

from google_code import google_authenticator_token
import keyring
import pexpect


import hmac, base64, struct, hashlib, time


def getpassword(service, username, alias="password"):
    """Get password from keychain"""

    password = keyring.get_password(service, username + "/" + alias)
    # print("test....",password, username, alias, service)
    while not password:
        # ask and save a password.
        password = getpass.getpass(alias + ": ")
        # print("test..22",password)
        if not password:
            print("Please enter a " + alias)
    return password


def gettermsize():
    """horrible non-portable hack to get the terminal size to transmit
    to the child process spawned by pexpect"""
    (rows, cols) = os.popen("stty size").read().split()  # works on Mac OS X, YMMV
    rows = int(rows)
    cols = int(cols)
    return (rows, cols)


def setpassword(service, username, password):
    """Save password in keychain"""

    if not keyring.get_password(service, username):
        print(
            "Successful login - saving password for user %s under keychain service '%s'"
            % (username, service)
        )
        keyring.set_password(service, username, password)


def mask_code(code):
    """Mask the middle part of the code with asterisks"""
    if len(code) <= 4:
        return code
    return code[:2] + '*' * (len(code) - 4) + code[-2:]


def ssh(username, host, keychainservice="ssh_py_default", port=22):
    """Automate sending password when the server has public key auth disabled"""

    # cmd = "/usr/bin/ssh %s@%s" % (username, host)
    cmd = "/usr/bin/ssh %s" % (host)
    if username != "":
        cmd = "/usr/bin/ssh %s@%s" % (username, host)

    if port != 22:
        cmd = cmd + f" -p${port}"

    print("Connecting cmd %s" % (cmd))
    child = pexpect.spawn(cmd)
    (rows, cols) = gettermsize()
    child.setwinsize(rows, cols)  # set the child to the size of the user's term

    # handle the host acceptance and password crap.
    verificationCode = None
    password = None
    while True:
        i = child.expect(
            [
                r"Are you sure you want to continue connecting (yes/no)?",
                r"assword:",
                r"Connection refused",
                r"Verification code:",
                r"\$",
                r">:",
                r"Last login:",
                r"Disconnected from",
                r"\[OTP Code\]:",
                r"Opt",
                r"Host]",
            ],
            timeout=30,
        )

        # 打印到目前为止接收到的所有输出
        # print("Output before login prompt:", child.before.decode(), i)
        # 打印登录提示后的输出
        # print("Output after login prompt:", child.after.decode())
        # print(i)
        if i == 0:
            # accept the host
            print("New server, accept the host...")
            child.sendline("yes")
        elif i == 7:
            print("[ssh %s@%s -p%s ]: fail login " % (username, host, port))
            sys.exit(0)
        elif i == 1:
            print("Sending password")
            password = getpassword(
                keychainservice, "%s@%s" % (username, host), alias="password"
            )
            child.sendline(password)
        elif i == 2:
            print("ssh connection refused,%s@%s:%s" % (username, host, port))
            sys.exit(0)
        elif i == 3 or i == 8:
            # google auth code
            verificationCode = getpassword(
                keychainservice, "%s@%s" % (username, host), alias="googleAuthCode"
            )
            # print("google code :", verificationCode, "---\n")
            code = str(google_authenticator_token(verificationCode))
            masked_verificationCode = mask_code(verificationCode)
            print("google code :", masked_verificationCode, ":", code)
            # print(f"Sending google222 auth code,{code},{len(code)}")
            child.sendline(code)
            # child.sendline("\n")

        # assume we see a shell prompt ending in $ to denote successful login:
        elif i == 4 or i == 5 or i == 6 or i == 9 or i == 10:
            print("[ssh %s@%s -p%s ]: successful login " % (username, host, port))
            if password is not None:
                setpassword(
                    keychainservice, ("%s@%s/password") % (username, host), password
                )
            if verificationCode is not None:
                setpassword(
                    keychainservice,
                    ("%s@%s/googleAuthCode") % (username, host),
                    verificationCode,
                )
            break

    # give control to the human.
    # child.sendline()
    child.interact()


if __name__ == "__main__":
    parser = optparse.OptionParser(
        usage="sshpass.py [options] <username@>host, \n keyring & pexpect required, should pip install pexpect & keyring"
    )

    parser.add_option(
        "-k",
        "--keychainservice",
        dest="keychainservice",
        help="Keychain service name to store password under",
        default="ssh_py_default",
    )
    parser.add_option("-p", "--port", dest="port", help="SSH port", default=22)

    (opts, args) = parser.parse_args()

    if len(args) != 1:
        parser.print_usage()
        sys.exit(1)

    host_str = args[0]
    if host_str.find("@") != -1:
        (username, host) = host_str.split("@")
    else:
        # username = os.getlogin() # default to username on the current host.
        username = ""
        host = host_str

    ssh(username, host, port=int(opts.port), keychainservice=opts.keychainservice)
