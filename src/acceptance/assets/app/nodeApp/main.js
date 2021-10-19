const app = require("./app");


const start = (port) => {
    try {
        app.listen(port, () => {
            console.log(`APP running at http://localhost:${port}`);
        });
    } catch (err) {
        console.error(err);
        process.exit();
    }
};

start(process.env.PORT || 8080)
