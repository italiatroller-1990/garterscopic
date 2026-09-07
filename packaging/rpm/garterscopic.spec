%global debug_package %{nil}

Name:           garterscopic
Version:        %{VERSION}
Release:        1%{?dist}
Summary:        Declarative Lightweight Static Site Generator
License:        MIT
URL:            https://github.com/italiatroller-1990/garterscopic
Source0:        %{url}/archive/v%{version}.tar.gz
BuildArch:      %{ARCH}
BuildRequires:  bash
Requires:       bash

%description
Garterscopic is a simple, fast static site generator that lets you
build websites using HTML, YAML, Markdown, and CSS - no Node.js required.

%prep
%setup -q -n garterscopic-v%{version}

%build
CGO_ENABLED=0 go build -ldflags="-s -w" -o garterscopic ./cmd/garterscopic

%install
%make_install
install -Dm755 garterscopic %{buildroot}%{_bindir}/garterscopic
install -Dm644 packaging/rpm/garterscopic.bashcomp %{buildroot}%{_datadir}/bash-completion/completions/garterscopic
install -Dm644 packaging/rpm/garterscopic.zshcomp %{buildroot}%{_datadir}/zsh/site-functions/_garterscopic
install -Dm644 LICENSE %{buildroot}%{_docdir}/%{name}/LICENSE
install -Dm644 README.md %{buildroot}%{_docdir}/%{name}/README.md

%files
%{_bindir}/garterscopic
%{_datadir}/bash-completion/completions/garterscopic
%{_datadir}/zsh/site-functions/_garterscopic
%{_docdir}/%{name}/LICENSE
%{_docdir}/%{name}/README.md
%{_mandir}/man1/garterscopic.1*

%post
if [ -x /usr/bin/update-alternatives ]; then
    update-alternatives --install /usr/bin/garterscopic garterscopic %{_bindir}/garterscopic 100 2>/dev/null || true
fi

%preun
if [ $1 -eq 0 ]; then
    if [ -x /usr/bin/update-alternatives ]; then
        update-alternatives --remove garterscopic %{_bindir}/garterscopic 2>/dev/null || true
    fi
fi

%changelog
* $(date '+%a %b %d %Y') Garterscopic Team <italiatroller@protonmail.com> - %{version}
- See https://github.com/italiatroller-1990/garterscopic/releases
