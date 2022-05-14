<!DOCTYPE html>
<html>
<body>
<table class="head">
  <tr>
    <td class="head-ltitle">JUSTGREP(1)</td>
    <td class="head-vol">justgrep IRC tools</td>
    <td class="head-rtitle">JUSTGREP(1)</td>
  </tr>
</table>
<div class="manual-text">
<h1 class="Sh" title="Sh" id="NAME"><a class="permalink" href="#NAME">NAME</a></h1>
justgrep - Tool for scanning justlog logs
<h1 class="Sh" title="Sh" id="SYNOPSIS"><a class="permalink" href="#SYNOPSIS">SYNOPSIS</a></h1>
<b>justgrep</b> <i>[options]</i> <b>-channel</b> <i>channel name</i> <b>-url</b>
  <i>https://example.com</i> <b>-regex</b> <i>regular expression</i>
  <b>-start</b> <i>2021-01-01T00:00:00Z</i> [<b>-end</b>
  <i>2021-02-01T00:00:00Z</i>]
<div class="Pp"></div>
<div>&#x00A0;</div>
<b>justgrep</b> <i>[options]</i> <b>-r</b> <b>-url</b>
  <i>https://example.com</i> <b>-regex</b> <i>regular expression</i>
  <b>-start</b> <i>2021-01-01T00:00:00Z</i> [ <b>-end</b>
  <i>2021-02-01T00:00:00Z</i>]
<div class="Pp"></div>
<h1 class="Sh" title="Sh" id="DESCRIPTION"><a class="permalink" href="#DESCRIPTION">DESCRIPTION</a></h1>
This tool searches the desired <i>justlog instance</i> for a regular expression
  or username regular expression in a set time range.
<div class="Pp"></div>
<h1 class="Sh" title="Sh" id="OPTIONS"><a class="permalink" href="#OPTIONS">OPTIONS</a></h1>
<dl class="Bl-tag">
  <dt><b>-channel&#x00A0;</b>channel&#x00A0;name</dt>
  <dd>Pick desired channel to search.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-r</b></dt>
  <dd>Run search on all channels available on the desired <i>justlog
      instance</i>. Overrides <i>-channel</i>.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-max&#x00A0;</b>count</dt>
  <dd>Choose how many messages should be returned by <b>justgrep</b>.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-start</b>, <b>-end&#x00A0;</b>TIME</dt>
  <dd>Allow you to specify the time range to search. <i>-end</i> should be the
      later part of the range. <i>-end</i> defaults to the current date/time if
      not given. <i>-start</i> is required. Accepted formats are:
    <div class="Pp"></div>
    <table class="tbl">
      <tr>
        <td>1</td>
        <td> 2006-01-02 15:04:05</td>
      </tr>
      <tr>
        <td>2</td>
        <td> 2006-01-02 15:04:05-07:00</td>
      </tr>
      <tr>
        <td>3</td>
        <td> 2006-01-02T15:04:05Z07:00 (RFC3339)</td>
      </tr>
    </table>
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-user&#x00A0;</b>name</dt>
  <dd>Search logs for a single user. If <i>-uregex</i> is used in combination,
      <b>name</b> is treated as a regular expression. It's worth noting that
      search a single user's logs is much faster than a whole channel.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-notuser&#x00A0;</b>name</dt>
  <dd>Ignores user identified by <i>name</i> from log searches. If
      <i>-uregex</i> is used in combination, <b>name</b> is treated as a regular
      expression.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-uregex</b></dt>
  <dd>Switches <i>-user</i> and <i>-notuser</i> to be treated as a regular
      expression instead of literally.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-regex&#x00A0;</b>regular&#x00A0;expression</dt>
  <dd>Searches messages for the pattern. This option is required.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-url&#x00A0;</b>justlog&#x00A0;instance&#x00A0;url</dt>
  <dd>Selects your desired justlog instance. By default it's
      <i>http://localhost:8025</i>, the default listen address for justlog.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-v</b></dt>
  <dd>Shows you progress info on stderr. Not allowed with <i>-progress-json</i>.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-progress-json</b></dt>
  <dd>Returns the same information as <i>-v</i> but in JSON format for machine
      processing. Also uses stderr. Not allowed with <i>-v</i>.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-msg-only</b></dt>
  <dd>Deprecated: use <b>-msg-types PRIVMSG</b> instead.</dd>
  <dd>Makes <b>justgrep</b> return only user chat messages, <i>PRIVMSG</i>s.
    <div class="Pp"></div>
  </dd>
</dl>
<dl class="Bl-tag">
  <dt><b>-msg-types</b></dt>
  <dd>Makes justgrep return only certain messages based on the IRC command/action. Putting the most command types first might speed up your search slightly.
    <div class="Pp"></div>
  </dd>
</dl>
<h1 class="Sh" title="Sh" id="EXAMPLES"><a class="permalink" href="#EXAMPLES">EXAMPLES</a></h1>
Fetch all messages matching <i>pajaS</i> from <i>2021-12-01</i> to
  <i>2021-12-07</i> (inclusive) from channel <i>pajlada</i> from <i>justlog
  instance</i>:
<div class="Pp"></div>
<br/>
<pre>
justgrep -channel pajlada -regex &quot;pajaS&quot; -start 2021-12-01T00:00:00Z -end 2021-12-07T23:59:59Z -url [justlog instance]
</pre>
<br/>
<div class="Pp"></div>
Fetch all messages matching <i>pajaS</i> from <i>2021-12-01</i> until now from
  channel <i>pajlada</i> from <i>justlog instance</i>:
<div class="Pp"></div>
<br/>
<pre>
justgrep -channel pajlada -regex &quot;pajaS&quot; -start 2021-12-01T00:00:00Z -url [justlog instance]
</pre>
<br/>
Fetch all timeouts matching from <i>2021-12-01</i> to <i>2021-12-07</i>
  (inclusive) from channel <i>pajlada</i> from <i>justlog instance</i>:
<div class="Pp"></div>
<br/>
<pre>
justgrep -channel pajlada -msg-types CLEARCHAT -start 2021-12-01T00:00:00Z -end 2021-12-07T23:59:59Z -url [justlog instance]
</pre>
<br/>
<div class="Pp"></div>
<h1 class="Sh" title="Sh" id="SEE_ALSO"><a class="permalink" href="#SEE_ALSO">SEE
  ALSO</a></h1>
<b>irc2json</b>(1)</div>
<table class="foot">
  <tr>
    <td class="foot-date">2021-12-23</td>
    <td class="foot-os">Mm2PL</td>
  </tr>
</table>
</body>
</html>
