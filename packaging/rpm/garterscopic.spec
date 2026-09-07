%global debug_package %{nil}

Name:           garterscopic
Version:        %{_version}
Release:        1%{?dist}
Summary:        Declarative Lightweight Static Site Generator
License:        Apache-2.0
URL:            https://github.com/italiatroller-1990/garterscopic
BuildArch:      %{_arch}

%description
Garterscopic is a simple, fast static site generator that lets you
build websites using HTML, YAML, Markdown, and CSS - no Node.js required.

%install
install -Dm755 %{_sourcedir}/garterscopic %{buildroot}%{_bindir}/garterscopic
install -Dm644 %{_sourcedir}/LICENSE %{buildroot}%{_docdir}/%{name}/LICENSE
install -Dm644 %{_sourcedir}/README.md %{buildroot}%{_docdir}/%{name}/README.md
install -Dm644 %{_sourcedir}/packaging/rpm/garterscopic.bashcomp %{buildroot}%{_datadir}/bash-completion/completions/garterscopic
install -Dm644 %{_sourcedir}/packaging/rpm/garterscopic.zshcomp %{buildroot}%{_datadir}/zsh/site-functions/_garterscopic

%files
%{_bindir}/garterscopic
%{_datadir}/bash-completion/completions/garterscopic
%{_datadir}/zsh/site-functions/_garterscopic
%{_docdir}/%{name}/LICENSE
%{_docdir}/%{name}/README.md

%changelog
* Sat Sep 07 2026 Garterscopic Team <italiatroller@protonmail.com> - %{version}
- See https://github.com/italiatroller-1990/garterscopic/releases
