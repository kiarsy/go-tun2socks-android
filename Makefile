GOMOBILE=gomobile
GOBIND=$(GOMOBILE) bind
BUILDDIR=$(shell pwd)/build
ARTIFACT=$(BUILDDIR)/tun2socks.aar
LDFLAGS='-s -w'
IMPORT_PATH=https://github.com/kiarsy/go-tun2socks-android

BUILD_CMD="cd $(BUILDDIR) && $(GOBIND) -a -ldflags $(LDFLAGS) -target=android -tags android -o $(ARTIFACT) $(IMPORT_PATH)"
BUILD_DEBUG_CMD="cd $(BUILDDIR) && $(GOBIND) -a -ldflags $(LDFLAGS) -target=android -tags 'android debug' -o $(ARTIFACT) $(IMPORT_PATH)"

all: $(ARTIFACT)

$(ARTIFACT):
	mkdir -p $(BUILDDIR)
	eval $(BUILD_CMD)

debug:
	mkdir -p $(BUILDDIR)
	eval $(BUILD_DEBUG_CMD)

clean:
	rm -rf $(BUILDDIR)
