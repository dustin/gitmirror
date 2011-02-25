#! /bin/pfsh
#
# Software License Agreement (BSD License)
#
# Copyright (c) 2011, Trond Norbye
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are
# met:
#
#     * Redistributions of source code must retain the above copyright
# notice, this list of conditions and the following disclaimer.
#
#     * Redistributions in binary form must reproduce the above
# copyright notice, this list of conditions and the following disclaimer
# in the documentation and/or other materials provided with the
# distribution.
#
#     * The names of the contributors may be used to endorse or promote
# products derived from this software without specific prior written
# permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
# "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
# LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
# A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
# OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
# SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
# LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
# DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
#

PATH=/sbin:/usr/sbin:/bin:$PATH
cd `dirname $0`

usage()
{
  cat <<EOF

Usage $0 [-z pool] [-s] [-u] [-D zfspool]
   -D pool     Destroy the zfs filesystems in pool
   -z pool     ZFS pool to create the file sets in
   -s          Install SMF script
   -u          Define group, role, auths and profile

EOF
  exit 1
}

create_users()
{
   #
   # Create authorizations if they're not defined
   #
   fgrep solaris.smf.value.gitmirror /etc/security/auth_attr > /dev/null
   if [ $? -ne 0 ]
   then
      ed /etc/security/auth_attr > /dev/null <<EOF
a
solaris.smf.value.gitmirror:::Change Gitmirror value properties::
solaris.smf.manage.gitmirror:::Manage Gitmirror service states::
.
w
q
EOF
      if [ $? -ne 0 ]
      then
         echo "Failed to add gitmirror authorization definitions"
         exit 1
      fi
   fi

   #
   # Create the profile if it's not already defined
   #
   fgrep solaris.smf.manage.gitmirror /etc/security/prof_attr > /dev/null
   if [ $? -ne 0 ]
   then
      ed /etc/security/prof_attr > /dev/null <<EOF
a
Membase Administration::::auths=solaris.smf.manage.gitmirror,solaris.smf.value.gitmirror
.
w
q
EOF

      if [ $? -ne 0 ]
      then
         echo "Failed to add gitmirror profile definitions"
         exit 1
      fi
   fi

   #
   # Define the group unless it's already defined
   #
   getent group gitmirror > /dev/null
   if [ $? -ne 0 ]
   then
      groupadd gitmirror
      if [ $? -ne 0 ]
      then
         echo "Failed to create group gitmirror"
         exit 1
      fi
   fi

   #
   # Define the user unless it's already defined
   #
   getent passwd gitmirror > /dev/null
   if [ $? -ne 0 ]
   then
      roleadd -c "gitmirror daemon" -d /var/opt/gitmirror -g gitmirror \
              -A solaris.smf.value.gitmirror,solaris.smf.manage.gitmirror gitmirror
      if [ $? -ne 0 ]
      then
          echo "Failed to create role gitmirror"
          exit 1
      fi
   fi
}

install_smf()
{
    # Install the smf files
    install -f /lib/svc/method gitmirror
    if [ $? -ne 0 ]
    then
        echo "Failed to install smf startup script"
        exit 1
    fi

    install -f /var/svc/manifest/application -m 0444 gitmirror.xml
    if [ $? -ne 0 ]
    then
        echo "Failed to install smf definition"
        exit 1
    fi

    svccfg import /var/svc/manifest/application/gitmirror.xml
    if [ $? -ne 0 ]
    then
        echo "Failed to import smf definition"
        exit 1
    fi
}

set -- `getopt D:z:suh $*`
if [ $? != 0 ]
then
   usage
fi

for i in $*
do
   case $i in
   -D)  zfs destroy -f -r $2/gitmirror
        shift 2
        ;;

   -z)  zfsroot=$2
        shift 2
        ;;
   -s)  setup_smf=yes
        shift
        ;;
   -u)  setup_users=yes
        shift
        ;;
    --) shift
        break
        ;;
   -h)  usage
        ;;
   esac
done

if test "x${setup_users}" = "xyes"
then
   create_users
fi

if test "x${setup_smf}" = "xyes"
then
   install_smf
fi

if test "x${zfsroot}" != "x"
then
   zfs create ${zfsroot}/gitmirror
   zfs create -o mountpoint=/etc/opt/gitmirror ${zfsroot}/gitmirror/etc
   zfs create -o mountpoint=/var/opt/gitmirror ${zfsroot}/gitmirror/var
   zfs create -o mountpoint=/opt/gitmirror ${zfsroot}/gitmirror/opt
else
   for f in /etc/opt/gitmirror \
            /var/opt/gitmirror \
            /opt/gitmirror
   do
      if ! test -d ${f}
      then
         mkdir -p $f
      fi
   done
fi

chown -R gitmirror:gitmirror /etc/opt/gitmirror 2> /dev/zero
chown -R gitmirror:gitmirror /var/opt/gitmirror 2> /dev/zero
chown -R gitmirror:gitmirror /opt/gitmirror 2> /dev/zero

install -f /opt/gitmirror -m 0444 -u bin -g bin \
        ../../gitmirror.js
