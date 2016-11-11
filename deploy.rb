#!/usr/bin/env ruby
# frozen_string_literal: true
#
# Copyright (C) 2016 Harald Sitter <sitter@kde.org>
#
# This program is free software; you can redistribute it and/or
# modify it under the terms of the GNU General Public License as
# published by the Free Software Foundation; either version 3 of
# the License or any later version accepted by the membership of
# KDE e.V. (or its successor approved by the membership of KDE
# e.V.), which shall act as a proxy defined in Section 14 of
# version 3 of the license.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

bin = 'neon-sftp-bridge'
user = 'neon-sftp-bridge'
host = 'drax'
path = "/home/#{user}/bin/"
systemdpath = "/home/#{user}/.config/systemd/user"

system('ssh', "#{user}@#{host}",
       "mkdir -p #{systemdpath}") || raise
system('rsync', '-avz', '--progress',
       '-e', 'ssh', "#{bin}.service", "#{user}@#{host}:#{systemdpath}") || raise

system('rsync', '-avz', '--progress',
       '-e', 'ssh', bin, "#{user}@#{host}:#{path}") || raise

# drax has no systemd QQ
system('ssh', "root@#{host}",
       "restart #{bin}") || raise

# system('ssh', "#{user}@#{host}",
#        'systemctl --user enable neon-contents-grapple.service') || raise
# system('ssh', "#{user}@#{host}",
#        'systemctl --user restart neon-contents-grapple.service') || raise
