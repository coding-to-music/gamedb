const $bundlePage = $('#bundle-page');

if ($bundlePage.length > 0) {

    $.ajax({
        type: "GET",
        url: '/bundles/' + $bundlePage.attr('data-id') + '/prices.json',
        dataType: 'json',
        success: function (data, textStatus, jqXHR) {

            if (data === null) {
                data = [];
            }

            Highcharts.chart('prices-chart', {
                chart: {
                    zoomType: 'x',
                    backgroundColor: 'rgba(0,0,0,0)',
                },
                title: {
                    text: ''
                },
                subtitle: {
                    text: ''
                },
                tooltip: {
                    formatter: function () {
                        return this.y.toLocaleString() + '% discount on ' + moment(this.x).format("dddd DD MMM YYYY @ HH:mm");
                    }
                },
                xAxis: {
                    title: {
                        text: 'Date'
                    },
                    type: 'datetime',
                    labels: {
                        step: 1,
                        formatter: function () {
                            return moment(this.value).format("Do MMM YY");
                        },
                    },
                },
                yAxis: {
                    title: {
                        text: 'Discount (%)'
                    },
                    type: 'linear',
                    max: 0,
                    min: -100,
                    allowDecimals: false,
                    reversed: true,
                },
                legend: {
                    enabled: false
                },
                credits: {
                    enabled: false
                },
                series: [
                    {
                        type: 'line',
                        name: 'Price',
                        step: 'left',
                        color: '#28a745',
                        data: data,
                    }
                ],
            });
        },
    });
}