<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8"/>
  <style>
    table.head, table.foot { width: 100%; }
    td.head-rtitle, td.foot-os { text-align: right; }
    td.head-vol { text-align: center; }
    div.Pp { margin: 1ex 0ex; }
    div.Nd, div.Bf, div.Op { display: inline; }
    span.Pa, span.Ad { font-style: italic; }
    span.Ms { font-weight: bold; }
    dl.Bl-diag > dt { font-weight: bold; }
    code.Nm, code.Fl, code.Cm, code.Ic, code.In, code.Fd, code.Fn,
    code.Cd { font-weight: bold; font-family: inherit; }
  </style>
  <title>IRC2JSON(1)</title>
</head>
<body>
<table class="head">
  <tr>
    <td class="head-ltitle">IRC2JSON(1)</td>
    <td class="head-vol">justgrep IRC tools</td>
    <td class="head-rtitle">IRC2JSON(1)</td>
  </tr>
</table>
<div class="manual-text">
<h1 class="Sh" title="Sh" id="NAME"><a class="permalink" href="#NAME">NAME</a></h1>
irc2json - converts from data IRC to JSON
<h1 class="Sh" title="Sh" id="SYNOPSIS"><a class="permalink" href="#SYNOPSIS">SYNOPSIS</a></h1>
cat ./file_with_irc_messages | <b>irc2json</b>
<div class="Pp"></div>
<h1 class="Sh" title="Sh" id="DESCRIPTION"><a class="permalink" href="#DESCRIPTION">DESCRIPTION</a></h1>
Usage of this tool is dead simple. It reads IRC from stdin and writes JSON on
  stdout. This tool outputs JSON in the following format:
<div class="Pp"></div>
<pre>
{
  &quot;raw&quot;: &quot;@key=value;key2=value;time=2021-12-23T21:53:20.471Z :Mm2PL!~mm2pl@kotmisia.pl PRIVMSG #botwars :This is a test message&quot;,
  &quot;prefix&quot;: &quot;Mm2PL!~mm2pl@kotmisia.pl&quot;,
  &quot;user&quot;: &quot;mm2pl&quot;,
  &quot;args&quot;: [
    &quot;#botwars&quot;,
    &quot;This is a test message&quot;
  ],
  &quot;action&quot;: &quot;PRIVMSG&quot;,
  &quot;tags&quot;: {
    &quot;key&quot;: &quot;value&quot;,
    &quot;key2&quot;: &quot;value&quot;
  },
  &quot;timestamp&quot;: &quot;2021-12-23T22:53:20.471+01:00&quot;
}
</pre>
<div class="Pp"></div>
It is worth noting that the timestamp field will automatically be set if either
  <i>tmi-sent-ts</i> or <i>time</i> tag is set. Otherwise it will be an empty
  string.
<div class="Pp"></div>
<h1 class="Sh" title="Sh" id="EXIT_CODES"><a class="permalink" href="#EXIT_CODES">EXIT
  CODES</a></h1>
This tool fails with exit code 1 if it is unable to parse given message or
  failed to serialize it.
<div class="Pp"></div>
<h1 class="Sh" title="Sh" id="EXAMPLES"><a class="permalink" href="#EXAMPLES">EXAMPLES</a></h1>
Parse the message from the description:
<div class="Pp"></div>
<br/>
<pre>
echo '@key=value;key2=value;time=2021-12-23T21:53:20.471Z :Mm2PL!~mm2pl@kotmisia.pl PRIVMSG #botwars :This is a test message' | irc2json
</pre>
<br/>
<div class="Pp"></div>
Fetch and parse logs from justlog with <b>justgrep</b>(1) from <i>justlog
  instance</i>:
<div class="Pp"></div>
<br/>
<pre>
justgrep -channel pajlada -regex &quot;pajaS&quot; -start 2021-12-01T00:00:00Z -end 2021-12-07T23:59:59Z -url  <i>justlog instance</i> | irc2json
</pre>
<br/>
<div class="Pp"></div>
<h1 class="Sh" title="Sh" id="SEE__ALSO"><a class="permalink" href="#SEE__ALSO">SEE&#x00A0;ALSO</a></h1>
<b>jq</b>(1) <b>justgrep</b>(1)</div>
<table class="foot">
  <tr>
    <td class="foot-date">2021-12-23</td>
    <td class="foot-os">Mm2PL</td>
  </tr>
</table>
</body>
</html>
