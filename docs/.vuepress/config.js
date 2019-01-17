module.exports = {
    title: "qbase",
    description: "Documentation for the qbase.",
    dest: "./dist/docs",
    base: "/docs/",
    markdown: {
        lineNumbers: true
    },
    themeConfig: {
        lastUpdated: "Last Updated",
        nav: [{text: "Back to qbase", link: "https://www.github.com/qbaseGroup/qbase"}],
        sidebar: [
            {
                title: "Introduction",
                collapsable: false,
                children: [
                    ["/introduction/qbase", "qbase"]
                ]
            },
            {
                title: "Getting Started",
                collapsable: false,
                children: [
                    ["/getting-started/quick_start", "Quick Start"]
                ]
            },
            {
                title: "Client",
                collapsable: false,
                children: [
                    ["/client/command", "command"]
                ]
            }
            ,
            {
                title: "Spec",
                collapsable: false,
                children: [
                    ["/spec/qcp", "QCP"],
                    ["/spec/gas", "Gas"],
                    ["/spec/transaction", "Transaction"]
                ]
            }
        ]
    }
}
