/etc/kubernetes/manifests/rescheduler.manifest:
  file.managed:
    - source: salt://rescheduler/rescheduler.manifest
    - template: jinja
    - user: root
    - group: root
    - mode: 644
    - makedirs: true
    - dir_mode: 755
